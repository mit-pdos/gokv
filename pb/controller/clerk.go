package controller

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
)

const CONTROLLER_ADD = uint64(0)

type ControllerClerk struct {
	cl *rpc.RPCClient
}

func (ck *ControllerClerk) AddNewServer(newServer grove_ffi.Address) {
	enc := marshal.NewEnc(8)
	enc.PutInt(newServer)
	raw_args := enc.Finish()
	reply := new([]byte)
	ck.cl.Call(CONTROLLER_ADD, raw_args, reply, 100 /* ms */)
}

func MakeControllerClerk(host grove_ffi.Address) *ControllerClerk {
	ck := new(ControllerClerk)
	ck.cl = rpc.MakeRPCClient(host)
	return ck
}
