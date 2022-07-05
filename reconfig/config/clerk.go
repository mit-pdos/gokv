package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
)

type Clerk struct {
}

type Error uint64

func MakeClerk(host grove_ffi.Address) *Clerk {
	panic("config: impl")
}

func (ck *Clerk) GetFreshEpochAndRead() (uint64, []byte) {
	panic("config: impl")
}

func (ck *Clerk) Write(epoch uint64, v []byte) Error {
	panic("config: impl")
}
