package pb

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
	RPC_APPLY         = uint64(0)
	RPC_SETSTATE      = uint64(1)
	RPC_GETSTATE      = uint64(2)
	RPC_BECOMEPRIMARY = uint64(4)
	RPC_PRIMARYAPPLY  = uint64(5)
)

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) Apply(args *ApplyArgs) e.Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_APPLY, EncodeApplyArgs(args), reply, 100 /* ms */)
	if err != 0 {
		return e.Timeout
	} else {
		return e.DecodeError(*reply)
	}
}

func (ck *Clerk) SetState(args *SetStateArgs) e.Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_SETSTATE, EncodeSetStateArgs(args), reply, 1000 /* ms */)
	if err != 0 {
		return e.Timeout
	} else {
		return e.DecodeError(*reply)
	}
}

func (ck *Clerk) GetState(args *GetStateArgs) *GetStateReply {
	reply := new([]byte)
	err := ck.cl.Call(RPC_GETSTATE, EncodeGetStateArgs(args), reply, 1000 /* ms */)
	if err != 0 {
		return &GetStateReply{Err: e.Timeout}
	} else {
		return DecodeGetStateReply(*reply)
	}
}

func (ck *Clerk) BecomePrimary(args *BecomePrimaryArgs) e.Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_BECOMEPRIMARY, EncodeBecomePrimaryArgs(args), reply, 100 /* ms */)
	if err != 0 {
		return e.Timeout
	} else {
		return e.DecodeError(*reply)
	}
}

func (ck *Clerk) PrimaryApply(op []byte) (e.Error, []byte) {
	reply := new([]byte)
	err := ck.cl.Call(RPC_PRIMARYAPPLY, op, reply, 200 /* ms */)
	if err == 0 {
		err, _ := marshal.ReadInt(*reply)
		return err, (*reply)[8:]
	} else {
		return err, nil
	}
}
