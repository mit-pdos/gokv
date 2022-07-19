package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type Clerk struct {
	cl *urpc.Client
}

const (
	RPC_GETEPOCH    = uint64(0)
	RPC_WRITECONFIG = uint64(1)
)

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) GetEpochAndConfig() (uint64, []grove_ffi.Address) {
	reply := new([]byte)
	for {
		err := ck.cl.Call(RPC_GETEPOCH, make([]byte, 0), reply, 100 /* ms */)
		if err == 0 {
			break
		} else {
			continue
		}
	}
	var epoch uint64
	epoch, *reply = marshal.ReadInt(*reply)
	config := DecodeConfig(*reply)
	return epoch, config
}

func (ck *Clerk) WriteConfig(epoch uint64, config []grove_ffi.Address) e.Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_GETEPOCH, make([]byte, 0), reply, 100 /* ms */)
	if err == 0 {
		e, _ := marshal.ReadInt(*reply)
		return e
	} else {
		return err
	}
}
