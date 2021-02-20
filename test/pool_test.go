package test

import (
	"fmt"
	liteminer "liteminer/pkg"
	"testing"
)

func TestClientChange(t *testing.T) {
	fmt.Println("starting client change test")
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
			t.Errorf("Received error %v when creating miner", err)
		}
		miners[i] = m
	}

	client := liteminer.CreateClient([]string{addr})
	client.Connect([]string{"aaaa"})

}

func TestBusyPool(t *testing.T) {
	fmt.Println("starting busy pool test")
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
			t.Errorf("Received error %v when creating miner", err)
		}
		miners[i] = m
	}

	client := liteminer.CreateClient([]string{addr})
	client2 := liteminer.CreateClient([]string{addr})

	data := "data"
	upperbound := uint64(100000)

	nonces, err := client.Mine(data, upperbound)
	fmt.Println("error 1", err)

	nonces2, err := client2.Mine(data, upperbound)
	fmt.Println("error 2", err)

	if err == nil {
		t.Errorf("Should have recieved pool connection error")
	}

	fmt.Println("Completed", nonces, nonces2)

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

	for _, nonce := range nonces {
		if nonce != int64(expected) {
			t.Errorf("Expected nonce %d, but received %d", expected, nonce)
		}
	}

}

//func TestBusyPool2(t *testing.T){
//	fmt.Println("starting busy pool 2 test")
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
//		m, err := liteminer.CreateMiner(addr)
//		if err != nil {
//			t.Errorf("Received error %v when creating miner", err)
//		}
//		miners[i] = m
//	}
//
//	client := liteminer.CreateClient([]string{addr})
//
//
//
//	data := "data"
//	upperbound := uint64(10000000000)
//
//	nonces, err := client.Mine(data, upperbound)
//
//	nonces2, err2 := client.Mine("test", 10000)
//	fmt.Println("test 1", nonces2 , err2)
//
//
//
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
//
//	for _, nonce := range nonces {
//		if nonce != int64(expected) {
//			t.Errorf("Expected nonce %d, but received %d", expected, nonce)
//		}
//	}
//
//}

func TestDuplicatePort(t *testing.T) {

	_, err := liteminer.CreatePool("1234")
	if err != nil {
		t.Errorf("Received error %v when creating pool", err)
	}
	_, err2 := liteminer.CreatePool("1234")

	if err2 == nil {
		t.Errorf("Expected Error Port Already in Use Recieved nil")
	}

}

func TestBadMessage(t *testing.T) {
	fmt.Println("starting busy pool 2 test")
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
			t.Errorf("Received error %v when creating miner", err)
		}
		miners[i] = m
	}

	client := liteminer.CreateClient([]string{addr})
	p = nil

	for _, v := range client.PoolConns {
		liteminer.SendMsg(v, liteminer.ErrorMsg("ERROR disconnect"))
	}

	data := "data"
	upperbound := uint64(1000)
	//liteminer.SendMsg(nil,io.EOF)

	nonces, _ := client.Mine(data, upperbound)
	fmt.Println(nonces)
	for _, v := range nonces {
		if v != -1 {
			t.Errorf("Expected -1 got %v", v)
		}
	}

}

//func TestPoolDisconnection(t *testing.T) {
//
//	p, err := liteminer.CreatePool("")
//	if err != nil {
//		t.Errorf("Received error %v when creating pool", err)
//	}
//
//	addr := p.Addr.String()
//
//	numMiners := 2
//	miners := make([]*liteminer.Miner, numMiners)
//
//	for i := 0; i < numMiners; i++ {
//		m, err := liteminer.CreateMiner(addr)
//		if err != nil {
//			t.Errorf("Received error %v when creating miner", err)
//		}
//		miners[i] = m
//	}
//
//	client := liteminer.CreateClient([]string{addr})
//
//	for _, v := range p.Miners {
//		v.Conn.Close()
//	}
//
//	//for _, conn := range client.PoolConns {
//	//	liteminer.SendMsg(conn, liteminer.ErrorMsg("error"))
//	//}
//
//	data := "data"
//	upperbound := uint64(1000)
//	//liteminer.SendMsg(nil,io.EOF)
//
//	nonces, _ := client.Mine(data, upperbound)
//	fmt.Println(nonces)
//	for _, v := range nonces {
//		if v != -1 {
//			t.Errorf("Expected -1 got %v", v)
//		}
//	}
//
//}
