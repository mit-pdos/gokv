package ctr

import (
	"log"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/erpc"
	"github.com/mit-pdos/gokv/fencing/ctr/error_gk"
	"github.com/mit-pdos/gokv/fencing/ctr/getreply_gk"
	"github.com/mit-pdos/gokv/fencing/ctr/putargs_gk"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
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
	valProph := primitive.NewProph()

	reply_ptr := new([]byte)
	err := c.cl.Call(RPC_GET, req, reply_ptr, 100 /* ms */)
	if err != 0 {
		log.Println("ctr: urpc get call failed/timed out")
		primitive.Exit(1)
	}
	r, _ := getreply_gk.Unmarshal(*reply_ptr)

	if r.Err != error_gk.ENone {
		log.Println("ctr: get() stale epoch number")
		primitive.Exit(1)
	}
	valProph.ResolveU64(r.Val)
	return r.Val
}

func (c *Clerk) Put(v uint64, epoch uint64) {
	args := &putargs_gk.S{
		V:     v,
		Epoch: epoch,
	}
	req := c.e.NewRequest(putargs_gk.Marshal(make([]byte, 0), *args))

	reply_ptr := new([]byte)
	err := c.cl.Call(RPC_PUT, req, reply_ptr, 100 /* ms */)
	if err != 0 {
		log.Println("ctr: urpc put call failed/timed out")
		primitive.Exit(1)
	}

	epochErr, _ := error_gk.Unmarshal(*reply_ptr)

	if epochErr != error_gk.ENone {
		log.Println("ctr: get() stale epoch number")
		primitive.Exit(1)
	}
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.cl = urpc.MakeClient(host)

	reply_ptr := new([]byte)
	err := ck.cl.Call(RPC_FRESHCID, make([]byte, 0), reply_ptr, 100 /* ms */)
	if err != 0 {
		// panic("ctr: urpc call failed/timed out")
		log.Println("ctr: urpc getcid call failed/timed out")
		primitive.Exit(1)
	}
	ck.e = erpc.MakeClient(marshal.NewDec(*reply_ptr).GetInt())

	return ck
}
