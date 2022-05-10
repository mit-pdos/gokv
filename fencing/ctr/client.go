package ctr

import (
	"github.com/mit-pdos/gokv/erpc"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
	"log"
)

const (
	RPC_GET      = uint64(0)
	RPC_PUT      = uint64(1)
	RPC_FRESHCID = uint64(2)
)

type Clerk struct {
	cl *urpc.Client
	e  *erpc.Client
}

func (c *Clerk) Get(epoch uint64) uint64 {
	enc := marshal.NewEnc(8)
	enc.PutInt(epoch)
	req := enc.Finish()
	valProph := machine.NewProph()

	reply_ptr := new([]byte)
	err := c.cl.Call(RPC_GET, req, reply_ptr, 100 /* ms */)
	if err != 0 {
		log.Println("ctr: urpc get call failed/timed out")
		grove_ffi.Exit(1)
	}
	r := DecGetReply(*reply_ptr)

	if r.err != ENone {
		log.Println("ctr: get() stale epoch number")
		grove_ffi.Exit(1)
	}
	valProph.ResolveU64(r.val)
	return r.val
}

func (c *Clerk) Put(v uint64, epoch uint64) {
	args := &PutArgs{
		v:     v,
		epoch: epoch,
	}
	req := c.e.NewRequest(EncPutArgs(args))

	reply_ptr := new([]byte)
	err := c.cl.Call(RPC_PUT, req, reply_ptr, 100 /* ms */)
	if err != 0 {
		log.Println("ctr: urpc put call failed/timed out")
		grove_ffi.Exit(1)
	}

	dec := marshal.NewDec(*reply_ptr)
	epochErr := dec.GetInt()

	if epochErr != ENone {
		log.Println("ctr: get() stale epoch number")
		grove_ffi.Exit(1)
	}
	return
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.cl = urpc.MakeClient(host)

	reply_ptr := new([]byte)
	err := ck.cl.Call(RPC_FRESHCID, make([]byte, 0), reply_ptr, 100 /* ms */)
	if err != 0 {
		// panic("ctr: urpc call failed/timed out")
		log.Println("ctr: urpc getcid call failed/timed out")
		grove_ffi.Exit(1)
	}
	ck.e = erpc.MakeClient(marshal.NewDec(*reply_ptr).GetInt())

	return ck
}
