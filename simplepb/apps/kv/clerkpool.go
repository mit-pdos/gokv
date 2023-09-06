package kv

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/kv"
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

func (ck *ClerkPool) Put(key, val string) {
	ck.doWithClerk(func(ck *Clerk) {
		ck.Put(key, val)
	})
}

func (ck *ClerkPool) Get(key string) string {
	var ret string
	ck.doWithClerk(func(ck *Clerk) {
		ret = ck.Get(key)
	})
	return ret
}

func (ck *ClerkPool) CondPut(key, expect, val string) string {
	var ret string
	ck.doWithClerk(func(ck *Clerk) {
		ret = ck.CondPut(key, expect, val)
	})
	return ret
}

func MakeKv(confHost grove_ffi.Address) *kv.Kv {
	ck := MakeClerkPool(confHost)
	return &kv.Kv{
		Put:            ck.Put,
		Get:            ck.Get,
		ConditionalPut: ck.CondPut,
	}
}
