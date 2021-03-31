package bench

import (
	"github.com/mit-pdos/gokv/goosekv"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"sync"
)

type GooseKVClerkPool struct {
	mu *sync.Mutex
	// queue of free clerks
	freeClerks []*goosekv.GoKVClerk
	numClerks  uint64
	cls        []*grove_ffi.RPCClient
}

// the hope is that after a while, the number of clerks needed to maintain a
// request rate for an open system benchmark will stabilize.
func (p *GooseKVClerkPool) Put(key uint64, value []byte) {
	p.mu.Lock()
	var ck *goosekv.GoKVClerk
	if len(p.freeClerks) == 0 {
		ck = goosekv.MakeKVClerkWithRPCClient(p.numClerks, p.cls[p.numClerks%uint64(len(p.cls))])
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

func MakeGooseKVClerkPool(numInit uint64, numClients uint64) *GooseKVClerkPool {
	p := new(GooseKVClerkPool)
	p.mu = new(sync.Mutex)
	p.freeClerks = make([]*goosekv.GoKVClerk, numInit)
	p.cls = make([]*grove_ffi.RPCClient, numClients)
	for i := uint64(0); i < numClients; i++ {
		p.cls[i] = grove_ffi.MakeRPCClient("localhost:12345")
	}

	for i := uint64(0); i < numInit; i++ {
		p.freeClerks[i] = goosekv.MakeKVClerkWithRPCClient(i, p.cls[i%numClients])
	}
	p.numClerks = numInit
	return p
}
