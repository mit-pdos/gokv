package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/eesm"
	"github.com/mit-pdos/gokv/simplepb/apps/kv"
	"github.com/mit-pdos/gokv/simplepb/clerk"
)

type Clerk struct {
	cl *clerk.Clerk
}

func MakeClerk(confHost grove_ffi.Address) *Clerk {
	return &Clerk{cl: clerk.Make(confHost)}
}

func (ck *Clerk) Put(key []byte, val []byte) {
	putArgs := &kv.PutArgs{
		Key: key,
		Val: val,
	}
	ck.cl.Apply(kv.EncodePutArgs(putArgs))
}

func (ck *Clerk) Get(key []byte) []byte {
	return ck.cl.Apply(eesm.MakeRequest(kv.EncodeGetArgs(key)))
}
