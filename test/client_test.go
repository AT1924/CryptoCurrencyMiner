package test

import (
	"fmt"
	liteminer "liteminer/pkg"

	"testing"
)

func TestGetClient(t *testing.T) {

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
	client := liteminer.CreateClient([]string{addr})
	data := "data"
	upperbound := uint64(10000)

	nonces, err := client.Mine(data, upperbound)

	clientAddr := p.GetClient()

	for _, v := range client.PoolConns {

		if v.Conn.LocalAddr().String() != clientAddr.String() {
			t.Errorf("Error pool client address does not match client local address")
		}

	}

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

func TestClientPoolDisconnect(t *testing.T) {

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
	client := liteminer.CreateClient([]string{addr})
	data := "data"
	upperbound := uint64(10000)
	client.PoolConns = nil
	nonces, err := client.Mine(data, upperbound)
	fmt.Println("Completed", nonces)
	if err == nil {
		t.Errorf("Error expected but recieved nil")
	}

}
