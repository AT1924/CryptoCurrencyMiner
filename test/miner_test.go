package test

import (
	"fmt"
	liteminer "liteminer/pkg"
	"testing"
)

//func TestTooManyMiners(t *testing.T) {
//
//	p, err := liteminer.CreatePool("")
//	if err != nil {
//		t.Errorf("Received error %v when creating pool", err)
//	}
//
//	addr := p.Addr.String()
//
//	numMiners := 1001
//	miners := make([]*liteminer.Miner, numMiners)
//
//	// test with dead miners, miners that are shutdown
//	for i := 0; i < numMiners; i++ {
//		m, err := liteminer.CreateMiner(addr)
//
//		if err != nil {
//			t.Errorf("Received error %v when creating miner", err)
//		}
//		miners[i] = m
//	}
//
//	client := liteminer.CreateClient([]string{addr})
//
//	data := "data"
//	upperbound := uint64(1000)
//	// when calling mine is when the program breaks
//	nonces, err := client.Mine(data, upperbound)
//	fmt.Println("Completed", nonces)
//
//	fmt.Println("Calculating expected nonce...")
//	lowestHashValue := liteminer.Hash(data, 0)
//	var expected uint64 = 0
//	for i := 1; uint64(i) < upperbound; i++ {
//		hashValue := liteminer.Hash(data, uint64(i))
//		if hashValue < lowestHashValue {
//			lowestHashValue = hashValue
//			expected = uint64(i)
//		}
//	}
//	fmt.Println("Expected nonce: ", expected)
//
//	if err != nil {
//		t.Errorf("Received error %v when mining", err)
//	} else {
//		for _, nonce := range nonces {
//			if nonce != int64(expected) {
//				t.Errorf("Expected nonce %d, but received %d", expected, nonce)
//			}
//		}
//	}
//}

func TestMinerPoolConnectionError(t *testing.T) {

	p, err := liteminer.CreatePool("")
	if err != nil {
		t.Errorf("Received error %v when creating pool", err)
	}

	addr := p.Addr.String()
	fakeADDR := ""
	numMiners := 10
	miners := make([]*liteminer.Miner, numMiners)
	fmt.Println("Real addr: ", addr, " fake addr:  ", fakeADDR)

	// test with dead miners, miners that are shutdown
	for i := 0; i < numMiners; i++ {

		m, err := liteminer.CreateMiner(fakeADDR)

		if err == nil {
			t.Errorf("Should have received missing address when connecting to pool instead recieved nil")
		}
		miners[i] = m
	}

}

func TestShutdownMinerError(t *testing.T) {

	p, err := liteminer.CreatePool("")
	if err != nil {
		t.Errorf("Received error %v when creating pool", err)
	}

	addr := p.Addr.String()

	numMiners := 10
	miners := make([]*liteminer.Miner, numMiners)

	// test with dead miners, miners that are shutdown
	for i := 0; i < numMiners; i++ {

		m, err := liteminer.CreateMiner(addr)

		if err != nil {
			t.Errorf("Recieved error when creating miner: %v", err)
		}
		miners[i] = m
	}

	shutdown := miners[0].IsShutdown
	fmt.Println(shutdown)
	if shutdown.Load() != false {
		t.Errorf("Miner should not be shutdown")
	}
	miners[0].Shutdown()
	shutdown = miners[0].IsShutdown
	if shutdown.Load() != true {
		t.Errorf("Miner should be shutdown")
	}

}

func TestMinerPoolShutdownError(t *testing.T) {

	p, err := liteminer.CreatePool("")
	if err != nil {
		t.Errorf("Received error %v when creating pool", err)
	}

	addr := p.Addr.String()

	numMiners := 10
	miners := make([]*liteminer.Miner, numMiners)

	// test with dead miners, miners that are shutdown
	for i := 0; i < numMiners; i++ {

		m, err := liteminer.CreateMiner(addr)

		if err != nil {
			t.Errorf("Recieved error when creating miner: %v", err)
		}
		miners[i] = m
	}
	client := liteminer.CreateClient([]string{addr})

	miners[7].Shutdown()
	data := "data"
	upperbound := uint64(100000)

	nonces, err := client.Mine(data, upperbound)
	fmt.Println("Completed", nonces)

	fmt.Println("Calculating expected nonce...")
	lowestHashValue := liteminer.Hash(data, 0)
	var expected uint64 = 0
	for i := 1; uint64(i) < upperbound; i++ {
		hashValue := liteminer.Hash(data, uint64(i))
		if hashValue < lowestHashValue {
			lowestHashValue = hashValue
			expected = uint64(i)
		}
	}
	fmt.Println("Expected nonce: ", expected)

	if err != nil {
		t.Errorf("Received error %v when mining", err)
	} else {
		for _, nonce := range nonces {
			if nonce != int64(expected) {
				t.Errorf("Expected nonce %d, but received %d", expected, nonce)
			}
		}
	}

}

func TestGetMiners(t *testing.T) {

	p, err := liteminer.CreatePool("")
	if err != nil {
		t.Errorf("Received error %v when creating pool", err)
	}

	addr := p.Addr.String()

	numMiners := 10
	miners := make([]*liteminer.Miner, numMiners)

	for i := 0; i < numMiners; i++ {

		m, err := liteminer.CreateMiner(addr)

		if err != nil {
			t.Errorf("Recieved error when creating miner: %v", err)
		}
		miners[i] = m
	}

	poolMiners := p.GetMiners()

	if len(poolMiners) != numMiners-1 {
		t.Errorf("Error length of get Miners not equal to numMiners")
	}

}

//func TestBadMinerMessage(t *testing.T) {
//
//	p, err := liteminer.CreatePool("")
//	if err != nil {
//		t.Errorf("Received error %v when creating pool", err)
//	}
//
//	addr := p.Addr.String()
//
//	numMiners := 10
//	miners := make([]*liteminer.Miner, numMiners)
//
//	for i := 0; i < numMiners; i++ {
//
//		m, err := liteminer.CreateMiner(addr)
//
//		if err != nil {
//			t.Errorf("Recieved error when creating miner: %v", err)
//		}
//		miners[i] = m
//	}
//
//	for _, miner := range p.Miners {
//
//		liteminer.SendMsg(miner, liteminer.ErrorMsg("Error"))
//
//	}
//
//	//client := liteminer.CreateClient([]string{addr})
//	//
//	//
//	//for _,v := range {
//	//	liteminer.SendMsg(v, liteminer.ErrorMsg("Error"))
//	//}
//
//}

func TestLostPoolConnection(t *testing.T) {

	p, err := liteminer.CreatePool("")
	if err != nil {
		t.Errorf("Received error %v when creating pool", err)
	}

	addr := p.Addr.String()

	numMiners := 2
	miners := make([]*liteminer.Miner, numMiners)

	for i := 0; i < numMiners; i++ {

		m, err := liteminer.CreateMiner(addr)

		if err != nil {
			t.Errorf("Recieved error when creating miner: %v", err)
		}
		miners[i] = m
	}

	poolMiners := p.GetMiners()

	if len(poolMiners) != numMiners-1 {
		t.Errorf("Error length of get Miners not equal to numMiners")
	}

}
