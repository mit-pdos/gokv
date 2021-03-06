package client

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/state"
	"log"
)

type Clerk struct {
	confCk    *config.Clerk
	primaryCk *state.Clerk
}

func Make(confHost grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.confCk = config.MakeClerk(confHost)
	for {
		config := ck.confCk.GetConfig()
		if len(config) == 0 {
			continue
		} else {
			ck.primaryCk = state.MakeClerk(config[0])
			break
		}
	}
	return ck
}

func (ck *Clerk) FetchAndAppend(key uint64, val []byte) []byte {
	var ret []byte
	for {
		var err e.Error
		err, ret = ck.primaryCk.FetchAndAppend(key, val)
		if err == e.None {
			grove_ffi.Sleep(uint64(100_000_000))
			break
		} else {
			log.Println("Error: ", err)
			config := ck.confCk.GetConfig()
			ck.primaryCk = state.MakeClerk(config[0])
			grove_ffi.Sleep(uint64(100_000_000))
			continue
		}
	}
	return ret
}
