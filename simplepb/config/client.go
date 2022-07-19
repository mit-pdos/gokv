package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) GetEpochAndConfig() (uint64, []grove_ffi.Address) {
	panic("")
}

func (ck *Clerk) WriteConfig(epoch uint64, config []grove_ffi.Address) e.Error {
	panic("")
}
