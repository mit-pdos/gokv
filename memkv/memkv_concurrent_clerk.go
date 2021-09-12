package memkv

import (
	"sync"
	"github.com/goose-lang/std"
)

type MemKVClerkPtr *MemKVClerk

type KVClerkPool struct {
	mu *sync.Mutex
	// queue of free clerks
	freeClerks []MemKVClerkPtr
	coord HostName
}

func (p *KVClerkPool) getClerk() *MemKVClerk {
	p.mu.Lock()
	if len(p.freeClerks) == 0 {
		p.mu.Unlock() // don't want to hold lock while making a fresh clerk
		return MakeMemKVClerk(p.coord)
	} else {
		ck := p.freeClerks[0]
		p.freeClerks = p.freeClerks[1:]
		p.mu.Unlock()
		return ck
	}
}

func (p *KVClerkPool) putClerk(ck *MemKVClerk) {
	// asynchronously put it back in the free list
	go func() {
		p.mu.Lock()
		p.freeClerks = append(p.freeClerks, ck)
		p.mu.Unlock()
	}()
}

// the hope is that after a while, the number of clerks needed to maintain a
// request rate for an open system benchmark will stabilize.
func (p *KVClerkPool) Put(key uint64, value []byte) {
	ck := p.getClerk()

	// we now own ck
	ck.Put(key, value)

	// done with ck, so asynchronously put it back in the free list
	p.putClerk(ck)
}

func (p *KVClerkPool) Get(key uint64) []byte {
	ck := p.getClerk()

	// we now own ck
	value := ck.Get(key)

	p.putClerk(ck)

	return value
}

// returns a slice of "values" (which are byte slices) in the same order as the
// keys passed in as input
// FIXME: benchmark
func (p *KVClerkPool) MGet(keys []uint64) [][]byte {
	vals := make([][]byte, len(keys))
	std.Multipar(uint64(len(keys)), func(i uint64) {
		vals[i] = p.Get(keys[i])
	})
	return vals
}

func MakeKVClerkPool(numInit uint64, coord HostName) *KVClerkPool {
	p := new(KVClerkPool)
	p.mu = new(sync.Mutex)
	p.coord = coord
	p.freeClerks = make([]MemKVClerkPtr, numInit)
	for i := uint64(0); i < numInit; i++ {
		p.freeClerks[i] = MakeMemKVClerk(p.coord)
	}

	return p
}
