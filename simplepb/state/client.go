package state

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

const (
	RPC_FAA = uint64(0)
)

type Clerk struct {
	cl *urpc.Client
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) FetchAndAppend(key uint64, val []byte) (e.Error, []byte) {
	var args = make([]byte, 0, 8+len(val))
	args = marshal.WriteInt(args, key)
	args = marshal.WriteBytes(args, val)
	reply := new([]byte)
	err := ck.cl.Call(RPC_FAA, args, reply, 200 /* ms */)
	if err == 0 {
		err, _ := marshal.ReadInt(*reply)
		return err, (*reply)[8:]
	} else {
		return err, nil
	}
}
