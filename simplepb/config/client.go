package config

import "github.com/mit-pdos/gokv/grove_ffi"

type Clerk struct {
}

func MakeClerk(grove_ffi.Address) *Clerk {
	panic("")
}

func (ck *Clerk) GetEpochAndConfig() (uint64, []grove_ffi.Address) {
	panic("")
}

func (ck *Clerk) WriteConfig(epoch uint64, config []grove_ffi.Address) uint64 {
	panic("")
}
