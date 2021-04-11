package fastkv

import (
	"github.com/upamanyus/urpc/rpc"
	"sync"
)

type GooseKVClerkPool struct {
	mu *sync.Mutex
	// queue of free clerks
	freeClerks []*GoKVClerk
	numClerks  uint64
	cls        []*rpc.RPCClient
}

// the hope is that after a while, the number of clerks needed to maintain a
// request rate for an open system benchmark will stabilize.
func (p *GooseKVClerkPool) Put(key uint64, value []byte) {
	p.mu.Lock()
	var ck *GoKVClerk
	if len(p.freeClerks) == 0 {
		ck = MakeKVClerkWithRPCClient(p.numClerks, p.cls[p.numClerks%uint64(len(p.cls))])
		p.numClerks++
	} else {
		ck = p.freeClerks[0]
		p.freeClerks = p.freeClerks[1:]
	}
	p.mu.Unlock()

	// we now own ck
	ck.Put(key, value)

	// done with ck, so asynchronously put it back in the free list
	go func() {
		p.mu.Lock()
		p.freeClerks = append(p.freeClerks, ck)
		p.mu.Unlock()
	}()
}

func (p *GooseKVClerkPool) Get(key uint64) []byte {
	p.mu.Lock()
	var ck *GoKVClerk
	if len(p.freeClerks) == 0 {
		ck = MakeKVClerkWithRPCClient(p.numClerks, p.cls[p.numClerks%uint64(len(p.cls))])
		p.numClerks++
	} else {
		ck = p.freeClerks[0]
		p.freeClerks = p.freeClerks[1:]
	}
	p.mu.Unlock()

	value := make([]byte, 0)
	var e ErrorType
	// we now own ck
	ck.Get(key, &e, &value)

	// done with ck, so asynchronously put it back in the free list
	go func() {
		p.mu.Lock()
		p.freeClerks = append(p.freeClerks, ck)
		p.mu.Unlock()
	}()
	return value
}

func MakeGooseKVClerkPool(numInit uint64, numClients uint64) *GooseKVClerkPool {
	p := new(GooseKVClerkPool)
	p.mu = new(sync.Mutex)
	p.freeClerks = make([]*GoKVClerk, numInit)
	p.cls = make([]*rpc.RPCClient, numClients)
	for i := uint64(0); i < numClients; i++ {
		p.cls[i] = rpc.MakeRPCClient("localhost:12345")
	}

	for i := uint64(0); i < numInit; i++ {
		p.freeClerks[i] = MakeKVClerkWithRPCClient(i, p.cls[i%numClients])
	}
	p.numClerks = numInit
	return p
}
