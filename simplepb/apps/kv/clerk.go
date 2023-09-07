package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/esm"
)

type Clerk struct {
	cl *esm.Clerk
}

func MakeClerk(confHosts []grove_ffi.Address) *Clerk {
	return &Clerk{cl: esm.MakeClerk(confHosts)}
}

func (ck *Clerk) Put(key, val string) {
	args := &PutArgs{
		Key: key,
		Val: val,
	}
	ck.cl.ApplyExactlyOnce(encodePutArgs(args))
}

func (ck *Clerk) Get(key string) string {
	return string(ck.cl.ApplyReadonly(encodeGetArgs(key)))
}

func (ck *Clerk) CondPut(key, expect, val string) string {
	args := &CondPutArgs{
		Key:    key,
		Expect: expect,
		Val:    val,
	}
	return string(ck.cl.ApplyExactlyOnce(encodeCondPutArgs(args)))
}
