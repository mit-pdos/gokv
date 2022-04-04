package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
	"log"
)

const (
	RPC_ACQUIRE_EPOCH = uint64(0)
	RPC_GET           = uint64(1)
	RPC_HB            = uint64(2)
)

type Clerk struct {
	cl *urpc.Client
}

func (ck *Clerk) HeartbeatThread(epoch uint64) {
	enc := marshal.NewEnc(8)
	enc.PutInt(epoch)
	args := enc.Finish()
	for {
		// XXX: make this statistically rigorous (e.g. aim for at most x% chance
		// of spurious leader failure per hour)
		reply_ptr := new([]byte)
		grove_ffi.Sleep(TIMEOUT_MS * MILLION / 3)
		err := ck.cl.Call(RPC_HB, args, reply_ptr, 100 /* ms */)
		if err != 0 || len(*reply_ptr) != 0 {
			break
		}
	}
}

func (ck *Clerk) AcquireEpoch(newFrontend grove_ffi.Address) uint64 {
	enc := marshal.NewEnc(8)
	enc.PutInt(newFrontend)
	reply_ptr := new([]byte)
	err := ck.cl.Call(RPC_ACQUIRE_EPOCH, enc.Finish(), reply_ptr, 100 /* ms */)
	if err != 0 {
		log.Println("config: client failed to run RPC on config server")
		grove_ffi.Exit(1)
	}
	dec := marshal.NewDec(*reply_ptr)
	return dec.GetInt()
}

func (ck *Clerk) Get() uint64 {
	reply_ptr := new([]byte)
	err := ck.cl.Call(RPC_GET, make([]byte, 0), reply_ptr, 100 /* ms */)
	if err != 0 {
		log.Println("config: client failed to run RPC on config server")
		grove_ffi.Exit(1)
	}
	dec := marshal.NewDec(*reply_ptr)
	return dec.GetInt()
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.cl = urpc.MakeClient(host)
	return ck
}
