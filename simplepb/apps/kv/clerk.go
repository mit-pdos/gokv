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
	putArgs := &putArgs{
		key:key,
		val:val,
	}
	ck.cl.Apply(encodePutArgs(putArgs));
}

func (ck *Clerk) Get(key []byte) []byte {
	return ck.cl.Apply(encodeGetArgs(key));
}
