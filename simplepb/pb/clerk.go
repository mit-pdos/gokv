package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

const (
	RPC_APPLY    = uint64(0)
	RPC_SETSTATE = uint64(1)
	RPC_GETSTATE = uint64(2)
)

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) Apply(args *ApplyArgs) Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_APPLY, EncodeApplyArgs(args), reply, 100 /* ms */)
	if err != 0 {
		return ETimeout
	} else {
		return DecodeError(*reply)
	}
}

func (ck *Clerk) SetState(args *SetStateArgs) Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_APPLY, EncodeSetStateArgs(args), reply, 1000 /* ms */)
	if err != 0 {
		return ETimeout
	} else {
		return DecodeError(*reply)
	}
}

func (ck *Clerk) GetState(args *GetStateArgs) *GetStateReply {
	reply := new([]byte)
	err := ck.cl.Call(RPC_APPLY, EncodeGetStateArgs(args), reply, 1000 /* ms */)
	if err != 0 {
		return &GetStateReply{Err: ETimeout}
	} else {
		return DecodeGetStateReply(*reply)
	}
}

func (ck *Clerk) BecomePrimary(args *BecomePrimaryArgs) Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_APPLY, EncodeBecomePrimaryArgs(args), reply, 100 /* ms */)
	if err != 0 {
		return ETimeout
	} else {
		return DecodeError(*reply)
	}
}
