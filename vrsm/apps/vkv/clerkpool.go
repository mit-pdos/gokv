package vkv

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/kv"
)

// TODO: should implement clerk pool at the exactlyonce level, so that we can
// avoid aking ownership of a whole clerk just to do a Get().
type ClerkPool struct {
	mu        *sync.Mutex
	cls       []*Clerk
	confHosts []grove_ffi.Address
}

func MakeClerkPool(confHosts []grove_ffi.Address) *ClerkPool {
	return &ClerkPool{
		mu:        new(sync.Mutex),
		cls:       make([]*Clerk, 0),
		confHosts: confHosts,
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
		cl = MakeClerk(ck.confHosts)

		f(cl)
		ck.mu.Lock()
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

func (ck *ClerkPool) ConditionalPut(key, expect, val string) string {
	var ret string
	ck.doWithClerk(func(ck *Clerk) {
		ret = ck.CondPut(key, expect, val)
	})
	return ret
}

func MakeKv(confHosts []grove_ffi.Address) kv.KvCput {
	ck := MakeClerkPool(confHosts)
	return ck
}
