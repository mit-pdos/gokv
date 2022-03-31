package frontend

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

const (
	RPC_FAI = uint64(0)
)

type Clerk struct {
	cl *urpc.Client
}

func (ck *Clerk) FetchAndIncrement(key uint64, ret *uint64) uint64 {
	reply_ptr := new([]byte)
	enc := marshal.NewEnc(8)
	enc.PutInt(key)
	err := ck.cl.Call(RPC_FAI, enc.Finish(), reply_ptr, 100 /* ms */)
	if err != 0 {
		return err
	}
	dec := marshal.NewDec(*reply_ptr)
	*ret = dec.GetInt()
	return 0
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.cl = urpc.MakeClient(host)
	return ck
}
