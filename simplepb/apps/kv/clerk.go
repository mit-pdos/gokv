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
	putArgs := &PutArgs{
		Key: key,
		Val: val,
	}
	ck.cl.Apply(EncodePutArgs(putArgs))
}

func (ck *Clerk) Get(key []byte) []byte {
	return ck.cl.ApplyRo(EncodeGetArgs(key))
}
