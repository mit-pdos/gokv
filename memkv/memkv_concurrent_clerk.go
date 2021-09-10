package memkv

import (
	"sync"
)

type KVClerk interface {
	Put(uint64, []byte)
	Get(uint64) []byte
}

type KVClerkPool struct {
	mu *sync.Mutex
	// queue of free clerks
	freeClerks []KVClerk
	factory    func() KVClerk
}

// the hope is that after a while, the number of clerks needed to maintain a
// request rate for an open system benchmark will stabilize.
func (p *KVClerkPool) Put(key uint64, value []byte) {
	p.mu.Lock()
	var ck KVClerk
	if len(p.freeClerks) == 0 {
		ck = p.factory()
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

func (p *KVClerkPool) Get(key uint64) []byte {
	p.mu.Lock()
	var ck KVClerk
	if len(p.freeClerks) == 0 {
		p.mu.Unlock() // don't want to hold lock while making a fresh clerk
		ck = p.factory()
	} else {
		ck = p.freeClerks[0]
		p.freeClerks = p.freeClerks[1:]
		p.mu.Unlock()
	}

	value := make([]byte, 0)
	// we now own ck
	value = ck.Get(key)

	p.mu.Lock()
	p.freeClerks = append(p.freeClerks, ck)
	p.mu.Unlock()
	return value
}

func MakeKVClerkPool(numInit uint64, numClients uint64, factory func() KVClerk) *KVClerkPool {
	p := new(KVClerkPool)
	p.mu = new(sync.Mutex)
	p.factory = factory
	p.freeClerks = make([]KVClerk, numInit)
	for i := uint64(0); i < numInit; i++ {
		p.freeClerks[i] = p.factory()
	}

	return p
}
