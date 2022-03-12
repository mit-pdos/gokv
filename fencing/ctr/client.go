package ctr

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
)

const (
	RPC_GET      = uint64(0)
	RPC_PUT      = uint64(1)
	RPC_FRESHCID = uint64(2)
)

type Clerk struct {
	cl  *rpc.RPCClient
	cid uint64
	seq uint64
}

func (c *Clerk) Get() uint64 {
	reply_ptr := new([]byte)
	err := c.cl.Call(RPC_GET, make([]byte, 0), reply_ptr, 100 /* ms */)
	if err != 0 {
		panic("ctr: urpc call failed/timed out")
	}
	dec := marshal.NewDec(*reply_ptr)
	return dec.GetInt()
}

func (c *Clerk) Put(v uint64) {
	reply_ptr := new([]byte)

	c.seq += 1
	args := &PutArgs{
		cid: c.cid,
		seq: c.seq,
		v:   v,
	}
	err := c.cl.Call(RPC_GET, EncPutArgs(args), reply_ptr, 100 /* ms */)
	if err != 0 {
		panic("ctr: urpc call failed/timed out")
	}
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.seq = 0
	ck.cl = rpc.MakeRPCClient(host)

	reply_ptr := new([]byte)
	err := ck.cl.Call(RPC_GET, make([]byte, 0), reply_ptr, 100 /* ms */)
	if err != 0 {
		panic("ctr: urpc call failed/timed out")
	}
	ck.cid = marshal.NewDec(*reply_ptr).GetInt()
	return ck
}
