package vkv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/vrsm/apps/exactlyonce"
	"github.com/mit-pdos/gokv/vrsm/apps/vkv/condputargs_gk"
	"github.com/mit-pdos/gokv/vrsm/apps/vkv/getargs_gk"
	"github.com/mit-pdos/gokv/vrsm/apps/vkv/putargs_gk"
)

type Clerk struct {
	cl *exactlyonce.Clerk
}

func MakeClerk(confHosts []grove_ffi.Address) *Clerk {
	return &Clerk{cl: exactlyonce.MakeClerk(confHosts)}
}

func (ck *Clerk) Put(key, val string) {
	args := putargs_gk.S{
		Key: key,
		Val: val,
	}
	ck.cl.ApplyExactlyOnce(putargs_gk.Marshal(make([]byte, 0), args))
}

func (ck *Clerk) Get(key string) string {
	args := getargs_gk.S{
		Get: key,
	}
	return string(ck.cl.ApplyReadonly(getargs_gk.Marshal(make([]byte, 0), args)))
}

func (ck *Clerk) CondPut(key, expect, val string) string {
	args := condputargs_gk.S{
		Key:    key,
		Expect: expect,
		Val:    val,
	}
	return string(ck.cl.ApplyExactlyOnce(condputargs_gk.Marshal(make([]byte, 0), args)))
}
