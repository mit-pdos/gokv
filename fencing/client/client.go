package client

import (
	"github.com/mit-pdos/gokv/fencing/config"
	"github.com/mit-pdos/gokv/fencing/frontend"
	"github.com/mit-pdos/gokv/grove_ffi"
)

type Clerk struct {
	configCk   *config.Clerk
	frontendCk *frontend.Clerk
}

func (ck *Clerk) FetchAndIncrement(key uint64) uint64 {
	ret := new(uint64)
	for {
		err := ck.frontendCk.FetchAndIncrement(key, ret)
		if err == 0 {
			break
		}
		currentFrontend := ck.configCk.Get()
		ck.frontendCk = frontend.MakeClerk(currentFrontend)
	}

	return *ret
}

func MakeClerk(configHost grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.configCk = config.MakeClerk(configHost)

	currentFrontend := ck.configCk.Get()
	ck.frontendCk = frontend.MakeClerk(currentFrontend)
	return ck
}
