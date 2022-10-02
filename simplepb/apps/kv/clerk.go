package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/clerk"
)

type Clerk struct {
	cl *clerk.Clerk
}

func MakeClerk(confHost grove_ffi.Address) *Clerk {
	return &Clerk{cl: clerk.Make(confHost)}
}

func (ck *Clerk) Put(key []byte, val []byte) {
	// ck.cl.Apply();
}

func (ck *Clerk) Get(key []byte) {

}
