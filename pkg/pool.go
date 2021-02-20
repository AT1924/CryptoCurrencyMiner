/*
 *  Brown University, CS138, Spring 2020
 *
 *  Purpose: a LiteMiner mining pool.
 */

package pkg

import (
	"encoding/gob"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"go.uber.org/atomic"
)

// HeartbeatTimeout is the time duration at which a pool considers a miner 'dead'
const HeartbeatTimeout = 3 * HeartbeatFreq

type AddressSpeedPair struct {
	addr  net.Addr
	speed float32
}

// Pool represents a LiteMiner mining pool
type Pool struct {
	Addr net.Addr

	Miners    map[net.Addr]MiningConn // Currently connected miners
	MinersMtx sync.Mutex              // Mutex for concurrent access to miners

	Client    MiningConn // The current client
	ClientMtx sync.Mutex // Mutex for concurrent access to Client

	busy *atomic.Bool // True when processing a transaction

	miningString *atomic.String

	MinerSpeeds      map[net.Addr]float32
	MinerSpeedsMutex sync.Mutex
	meanMinerSpeed   *atomic.Float64
	stddevMinerSpeed *atomic.Float64
	maxMinerSpeed    *atomic.Float64
	speedUpdatesChan chan AddressSpeedPair

	minersWait sync.WaitGroup

	numJobGoal      *atomic.Int64
	numJobsFinished *atomic.Int64

	jobsChan   chan Interval
	noncesChan chan Message
}

// CreatePool creates a new pool at the specified port.
func CreatePool(port string) (*Pool, error) {
	p := &Pool{
		busy:             atomic.NewBool(false),
		Miners:           make(map[net.Addr]MiningConn),
		speedUpdatesChan: make(chan AddressSpeedPair),
		MinerSpeeds:      make(map[net.Addr]float32),
		jobsChan:         make(chan Interval),
		miningString:     atomic.NewString(""),
		numJobsFinished:  atomic.NewInt64(0),
		meanMinerSpeed:   atomic.NewFloat64(0),
		stddevMinerSpeed: atomic.NewFloat64(0),
		maxMinerSpeed:    atomic.NewFloat64(0),
	}

	// TODO: Students should (if necessary) initialize any additional members
	// to the Pool struct here.

	err := p.startListener(port)

	return p, err
}

// startListener starts listening for new connections.
func (p *Pool) startListener(port string) error {
	listener, portID, err := OpenListener(port)
	if err != nil {
		return err
	}

	Out.Printf("Listening on port %v\n", portID)

	p.Addr = listener.Addr()

	// Listen for and accept connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				Err.Printf("Received error %v when listening for connections\n", err)
				continue
			}

			go p.handleConnection(conn)
		}
	}()

	return nil
}

// handleConnection handles an incoming connection and delegates to
// handleMinerConnection or handleClientConnection.
func (p *Pool) handleConnection(nc net.Conn) {
	// Set up connection
	conn := MiningConn{}
	conn.Conn = nc
	conn.Enc = gob.NewEncoder(nc)
	conn.Dec = gob.NewDecoder(nc)

	// Wait for Hello message
	msg, err := RecvMsg(conn)
	if err != nil {
		Err.Printf(
			"Received error %v when processing Hello message from %v\n",
			err,
			conn.Conn.RemoteAddr(),
		)
		conn.Conn.Close() // Close the connection
		return
	}

	switch msg.Type {
	case MinerHello:
		p.handleMinerConnection(conn)
	case ClientHello:
		p.handleClientConnection(conn)
	default:
		Err.Printf("Pool received unexpcted message type %v (msg=%v)", msg.Type, msg)
		SendMsg(conn, ErrorMsg("Unexpected message type"))
	}
}

// handleClientConnection handles a connection from a client.
func (p *Pool) handleClientConnection(conn MiningConn) {
	Debug.Printf("Received client connection from %v", conn.Conn.RemoteAddr())

	p.ClientMtx.Lock()
	if p.Client.Conn != nil {
		Debug.Printf(
			"Busy with client %v, sending BusyPool message to client %v",
			p.Client.Conn.RemoteAddr(),
			conn.Conn.RemoteAddr(),
		)
		SendMsg(conn, BusyPoolMsg())
		p.ClientMtx.Unlock()
		return
	}
	p.ClientMtx.Unlock()

	if p.busy.Load() {
		Debug.Printf(
			"Busy with previous transaction, sending BusyPool message to client %v",
			conn.Conn.RemoteAddr(),
		)
		SendMsg(conn, BusyPoolMsg())
		return
	}
	p.ClientMtx.Lock()
	p.Client = conn
	p.ClientMtx.Unlock()

	// Listen for and handle incoming messages
	for {
		msg, err := RecvMsg(conn)
		if err != nil {
			if _, ok := err.(*net.OpError); ok || err == io.EOF {
				Out.Printf("Client %v disconnected\n", conn.Conn.RemoteAddr())

				conn.Conn.Close() // Close the connection

				p.ClientMtx.Lock()
				p.Client.Conn = nil
				p.ClientMtx.Unlock()

				return
			}
			Err.Printf(
				"Received error %v when processing message from client %v\n",
				err,
				conn.Conn.RemoteAddr(),
			)
			return
		}

		if msg.Type != Transaction {
			SendMsg(conn, ErrorMsg("Expected Transaction message"))
			continue
		}

		Debug.Printf(
			"Received transaction from client %v with data %v and upper bound %v",
			conn.Conn.RemoteAddr(),
			msg.Data,
			msg.Upper,
		)

		p.MinersMtx.Lock()
		if len(p.Miners) == 0 {
			SendMsg(conn, ErrorMsg("No miners connected"))
			p.MinersMtx.Unlock()
			continue
		}
		p.MinersMtx.Unlock()

		// TODO: Students should handle an incoming transaction from a client. A
		// pool may process one transaction at a time – thus, if you receive
		// another transaction while busy, you should send a BusyPool message.
		// Otherwise, you should let the miners do their jobs. Note that miners
		// are handled in separate go routines (`handleMinerConnection`). To notify
		// the miners, consider using a shared data structure.
		//fmt.Println("CHECKING!")
		if p.busy.Load() {
			//fmt.Println("BUSY!!")
			SendMsg(conn, BusyPoolMsg())
		} else {
			p.busy.Store(true)
			if msg.Upper > 0 {
				//fmt.Println("message", msg.Upper)
				p.miningString.Store(msg.Data)
				//fmt.Println("miner", p.miningString)
				p.noncesChan = make(chan Message)
				p.MinersMtx.Lock()
				numMiners := len(p.Miners)
				p.MinersMtx.Unlock()
				// TODO: create a rule for number of intervals
				numInitialJobs := int(math.Min(float64(numMiners)*10, float64(msg.Upper)))
				intervalsSlice := GenerateIntervals(msg.Upper, numInitialJobs)
				// set the initial job number goal
				// this variable is used to close the nonce channel once mining has completed
				p.numJobGoal = atomic.NewInt64(int64(numInitialJobs))
				go p.allocateJobs(intervalsSlice)
				go p.findNonce(msg.Data, msg.Upper)
				go p.aggregateSpeeds()
			} else {
				p.ClientMtx.Lock()
				SendMsg(p.Client, ProofOfWorkMsg(p.miningString.Load(), 0, Hash(msg.Data, 0)))
				p.ClientMtx.Unlock()
				p.busy.Store(false)
			}
		}
	}
}

func (p *Pool) aggregateSpeeds() {
	for pair := range p.speedUpdatesChan {
		p.MinerSpeedsMutex.Lock()
		p.MinerSpeeds[pair.addr] = pair.speed
		total := float32(0)
		squaresTotal := float32(0)
		maxSpeed := 0.0
		for _, speed := range p.MinerSpeeds {
			total += speed
			squaresTotal += speed * speed
			if speed > float32(maxSpeed) {
				maxSpeed = float64(speed)
			}
		}
		p.maxMinerSpeed.Store(float64(maxSpeed))
		p.meanMinerSpeed.Store(float64(total / float32(len(p.MinerSpeeds))))
		squaresMean := squaresTotal / float32(len(p.MinerSpeeds))
		p.stddevMinerSpeed.Store(float64(math.Sqrt(float64(squaresMean) - p.meanMinerSpeed.Load()*p.meanMinerSpeed.Load())))
		p.maxMinerSpeed.Store(float64(maxSpeed))
		p.MinerSpeedsMutex.Unlock()
	}
}

func (p *Pool) allocateJobs(intervals []Interval) {
	for _, interval := range intervals {
		p.jobsChan <- interval
		//fmt.Println("hello!")
	}
}

func (p *Pool) findNonce(data string, upper uint64) {
	// find nonce from ProofOfWork messages in nonceChan, then send the nonce to the client
	// when noncesChan has been closed
	gotFirst := false
	var nonce uint64
	var smallestHash uint64
	for msg := range p.noncesChan {

		if !gotFirst {
			nonce = msg.Nonce
			smallestHash = msg.Hash
			gotFirst = true
		} else {
			if msg.Hash < smallestHash {
				nonce = msg.Nonce
				smallestHash = msg.Hash
			}
		}
	}
	if Hash(data, upper) < smallestHash {
		nonce = upper
	}
	p.ClientMtx.Lock()
	SendMsg(p.Client, ProofOfWorkMsg(p.miningString.Load(), nonce, smallestHash))
	p.ClientMtx.Unlock()
	p.busy.Store(false)
}

// handleMinerConnection handles a connection from a miner.
func (p *Pool) handleMinerConnection(conn MiningConn) {
	Debug.Printf("Received miner connection from %v", conn.Conn.RemoteAddr())

	p.MinersMtx.Lock()
	p.Miners[conn.Conn.RemoteAddr()] = conn
	p.MinersMtx.Unlock()

	msgChan := make(chan Message)
	go p.receiveFromMiner(conn, msgChan)

	// TODO: Students should handle a miner connection. If a miner does not
	// send a StatusUpdate message every `HeartbeatTimeout` while mining,
	// any work assigned to them should be redistributed and they should be
	// disconnected and removed from `p.Miners`.
	// For maintaining a queue of jobs yet to be taken, consider using a go channel.
	p.MinersMtx.Lock()
	addr := conn.Conn.RemoteAddr()
	p.MinersMtx.Unlock()
	savedTime := time.Now()
	for interval := range p.jobsChan {
		SendMsg(conn, MineRequestMsg(p.miningString.Load(), interval.Lower, interval.Upper))
		//fmt.Println("gave interval", interval.Lower, interval.Upper, " to ", addr)
		slowMiner := false
		gotProofOfWork := false
		timeout := time.After(time.Duration(HeartbeatTimeout))
		for !gotProofOfWork {
			select {
			case msg := <-msgChan:
				switch msg.Type {
				case StatusUpdate:
					secondsPassed := (time.Now().Sub(savedTime)).Seconds()
					if math.Abs(secondsPassed-1) > .05 {
						//fmt.Println("***********************************************")
						//fmt.Println("time: ", secondsPassed)
						savedTime = time.Now()
						break
					}
					numProcessed := msg.NumProcessed
					speed := float32(numProcessed) / float32(secondsPassed)
					var pair AddressSpeedPair
					pair.addr = addr
					pair.speed = speed
					p.speedUpdatesChan <- pair
					//fmt.Println("time: ", secondsPassed, " speed: ", speed, " maxspeed: ", p.maxMinerSpeed.Load(), "nums processed: ", numProcessed)
					if !slowMiner {
						if float64(speed)/p.maxMinerSpeed.Load() < .3 {
							//fmt.Println("found slow miner")
							slowMiner = true
							p.jobsChan <- interval
						}
					}
					savedTime = time.Now()
					break
				case ProofOfWork:
					gotProofOfWork = true
					if !slowMiner {
						p.noncesChan <- msg
						p.numJobsFinished.Add(1)
						//fmt.Println("Jobs: ", p.numJobsFinished.Load(), " out of ", p.numJobGoal)
						if p.numJobGoal.Load() == p.numJobsFinished.Load() {
							//fmt.Println("NONCES CHANNEL CLOSED")
							p.numJobGoal.Store(0)
							p.numJobsFinished.Store(0)
							close(p.noncesChan)
						}
					}
					break
				default:
					// TODO: raise error
					break
				}
			case <-timeout:
				p.MinerSpeedsMutex.Lock()
				delete(p.MinerSpeeds, addr)
				p.MinerSpeedsMutex.Unlock()
				p.MinersMtx.Lock()
				delete(p.Miners, addr)
				p.MinersMtx.Unlock()
				conn.Conn.Close()
				// reassign work that was given to this miner to another miner by placing the interval back in the jobs channel
				p.jobsChan <- interval
				return
			}
			timeout = time.After(time.Duration(HeartbeatTimeout))
		}
	}
}

// receiveFromMiner waits for messages from the miner specified by conn and
// forwards them over msgChan.
func (p *Pool) receiveFromMiner(conn MiningConn, msgChan chan Message) {
	for {
		msg, err := RecvMsg(conn)
		if err != nil {
			if _, ok := err.(*net.OpError); ok || err == io.EOF {
				Out.Printf("Miner %v disconnected\n", conn.Conn.RemoteAddr())

				p.MinersMtx.Lock()
				delete(p.Miners, conn.Conn.RemoteAddr())
				p.MinersMtx.Unlock()

				conn.Conn.Close() // Close the connection

				return
			}
			Err.Printf(
				"Received error %v when processing message from miner %v\n",
				err,
				conn.Conn.RemoteAddr(),
			)
			continue
		}
		msgChan <- msg
	}
}

// GetMiners returns the addresses of any connected miners.
func (p *Pool) GetMiners() []net.Addr {
	p.MinersMtx.Lock()
	defer p.MinersMtx.Unlock()

	miners := []net.Addr{}
	for _, m := range p.Miners {
		miners = append(miners, m.Conn.RemoteAddr())
	}
	return miners
}

// GetClient returns the address of the current client or nil if there is no
// current client.
func (p *Pool) GetClient() net.Addr {
	p.ClientMtx.Lock()
	defer p.ClientMtx.Unlock()

	if p.Client.Conn == nil {
		return nil
	}
	return p.Client.Conn.RemoteAddr()
}
