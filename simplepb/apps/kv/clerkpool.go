package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"sync"
)

type ClerkPool struct {
	mu       *sync.Mutex
	cls      []*Clerk
	confHost grove_ffi.Address
}

func MakeClerkPool(confHost grove_ffi.Address) *ClerkPool {
	return &ClerkPool{
		mu:       new(sync.Mutex),
		cls:      make([]*Clerk, 0),
		confHost: confHost,
	}
}

// TODO: get rid of stale clerks from the ck.cls list?
// TODO: keep failed clerks out of ck.cls list? Maybe f(cl) can return an
// optional error saying "get rid of cl".
// XXX: what's the performance overhead of function pointer here v.s. manually
// inlining the body each time?
func (ck *ClerkPool) doWithClerk(f func(ck *Clerk)) {
	ck.mu.Lock()
	var cl *Clerk
	if len(ck.cls) > 0 {
		// get the first cl
		cl = ck.cls[0]
		ck.cls = ck.cls[1:]
		ck.mu.Unlock()

		f(cl)
		// put cl back into the list
		ck.mu.Lock()
		ck.cls = append(ck.cls, cl)
		ck.mu.Unlock()

	} else {
		ck.mu.Unlock()
		cl = MakeClerk(ck.confHost)

		f(cl)
		// put the new cl into the list many times
		ck.mu.Lock()
		ck.cls = append(ck.cls, cl)
		ck.cls = append(ck.cls, cl)
		ck.cls = append(ck.cls, cl)
		ck.cls = append(ck.cls, cl)
		ck.mu.Unlock()
	}
}

func (ck *ClerkPool) Put(key []byte, val []byte) {
	ck.doWithClerk(func(ck *Clerk) {
		ck.Put(key, val)
	})
}

func (ck *ClerkPool) Get(key []byte) []byte {
	var ret []byte
	ck.doWithClerk(func(ck *Clerk) {
		ret = ck.Get(key)
	})
	return ret
}
