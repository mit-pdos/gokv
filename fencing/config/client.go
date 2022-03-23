package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"log"
)

const (
	RPC_LOCK = uint64(0)
	RPC_GET  = uint64(1)
)

type Clerk struct {
	cl *rpc.RPCClient
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
		err := ck.cl.Call(RPC_LOCK, args, reply_ptr, 100 /* ms */)
		if err != 0 || len(*reply_ptr) != 0 {
			break
		}
	}
}

func (ck *Clerk) Lock(newFrontend grove_ffi.Address) uint64 {
	enc := marshal.NewEnc(8)
	enc.PutInt(newFrontend)
	reply_ptr := new([]byte)
	err := ck.cl.Call(RPC_LOCK, enc.Finish(), reply_ptr, 100 /* ms */)
	if err != 0 {
		log.Fatalf("config: client failed to run RPC on config server")
	}
	dec := marshal.NewDec(*reply_ptr)
	return dec.GetInt()
}

func (ck *Clerk) Get() uint64 {
	reply_ptr := new([]byte)
	err := ck.cl.Call(RPC_LOCK, make([]byte, 0), reply_ptr, 100 /* ms */)
	if err != 0 {
		panic("config: client failed to run RPC on config server")
	}
	dec := marshal.NewDec(*reply_ptr)
	return dec.GetInt()
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.cl = rpc.MakeRPCClient(host)
	return ck
}
