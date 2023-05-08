package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/eesm"
)

type Clerk struct {
	cl *eesm.Clerk
}

func MakeClerk(confHost grove_ffi.Address) *Clerk {
	return &Clerk{cl: eesm.MakeClerk(confHost)}
}

func (ck *Clerk) Put(key []byte, val []byte) {
	putArgs := &PutArgs{
		Key: key,
		Val: val,
	}
	ck.cl.ApplyExactlyOnce(EncodePutArgs(putArgs))
}

func (ck *Clerk) Get(key []byte) []byte {
	return ck.cl.ApplyReadonly(EncodeGetArgs(key))
}
