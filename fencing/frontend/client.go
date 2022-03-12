package frontend

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
)

const (
	RPC_FAI = uint64(0)
)

type Clerk struct {
	cl *rpc.RPCClient
}

func (ck *Clerk) FetchAndIncrement(key uint64) uint64 {
	reply_ptr := new([]byte)
	enc := marshal.NewEnc(8)
	enc.PutInt(key)
	err := ck.cl.Call(RPC_FAI, enc.Finish(), reply_ptr, 100 /* ms */)
	if err != 0 {
		panic("disconnect")
	}
	dec := marshal.NewDec(*reply_ptr)
	return dec.GetInt()
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.cl = rpc.MakeRPCClient(host)
	return ck
}
