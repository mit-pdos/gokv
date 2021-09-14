package memkv

import (
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/connman"
	"sync"
)

type seqKVClerkPtr *SeqKVClerk

type KVClerk struct {
	mu *sync.Mutex
	// queue of free clerks
	freeClerks []seqKVClerkPtr
	cm         *connman.ConnMan
	coord      HostName
}

func (p *KVClerk) getSeqClerk() *SeqKVClerk {
	p.mu.Lock()
	n := len(p.freeClerks)
	if n == 0 {
		p.mu.Unlock() // don't want to hold lock while making a fresh clerk
		return MakeSeqKVClerk(p.coord, connman.MakeConnMan()) // XXX: different connman for each seqkvclerk
	} else {
		ck := p.freeClerks[n-1]
		p.freeClerks = p.freeClerks[:n-1]
		p.mu.Unlock()
		return ck
	}
}

func (p *KVClerk) putSeqClerk(ck *SeqKVClerk) {
	// asynchronously put it back in the free list
	go func() {
		p.mu.Lock()
		p.freeClerks = append(p.freeClerks, ck)
		p.mu.Unlock()
	}()
}

// the hope is that after a while, the number of clerks needed to maintain a
// request rate for an open system benchmark will stabilize.
func (p *KVClerk) Put(key uint64, value []byte) {
	ck := p.getSeqClerk()

	// we now own ck
	ck.Put(key, value)

	// done with ck, so asynchronously put it back in the free list
	p.putSeqClerk(ck)
}

func (p *KVClerk) Get(key uint64) []byte {
	ck := p.getSeqClerk()

	// we now own ck
	value := ck.Get(key)

	p.putSeqClerk(ck)

	return value
}

func (p *KVClerk) ConditionalPut(key uint64, expectedValue []byte, newValue []byte) bool {
	ck := p.getSeqClerk()

	// we now own ck
	ret := ck.ConditionalPut(key, expectedValue, newValue)

	// done with ck, so asynchronously put it back in the free list
	p.putSeqClerk(ck)

	return ret
}

func (p *KVClerk) Add(host HostName) {
	ck := p.getSeqClerk()
	ck.coordCk.AddShardServer(host)
	p.putSeqClerk(ck)
}

// returns a slice of "values" (which are byte slices) in the same order as the
// keys passed in as input
// FIXME: benchmark
func (p *KVClerk) MGet(keys []uint64) [][]byte {
	vals := make([][]byte, len(keys))
	std.Multipar(uint64(len(keys)), func(i uint64) {
		vals[i] = p.Get(keys[i])
	})
	return vals
}

func MakeKVClerk(coord HostName, cm *connman.ConnMan) *KVClerk {
	p := new(KVClerk)
	p.mu = new(sync.Mutex)
	p.coord = coord
	p.cm = cm
	p.freeClerks = make([]seqKVClerkPtr, 0)
	return p
}
