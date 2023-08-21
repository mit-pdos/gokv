package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/esm"
)

type Clerk struct {
	cl *esm.Clerk
}

func MakeClerk(confHost grove_ffi.Address) *Clerk {
	return &Clerk{cl: esm.MakeClerk(confHost)}
}

func (ck *Clerk) Put(key, val string) {
	putArgs := &PutArgs{
		Key: key,
		Val: val,
	}
	ck.cl.ApplyExactlyOnce(EncodePutArgs(putArgs))
}

func (ck *Clerk) Get(key string) string {
	return string(ck.cl.ApplyReadonly(EncodeGetArgs(key)))
}
