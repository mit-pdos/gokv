package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

const (
	RPCID_APPLY = uint64(0)
	RPCID_SETSTATE = uint64(1)
	RPCID_GETSTATE = uint64(2)
)

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) Apply(args *ApplyArgs) Error {
	panic("impl")
}

func (ck *Clerk) SetState(args *SetStateArgs) Error {
	panic("impl")
}

func (ck *Clerk) GetState(args *GetStateArgs) *GetStateReply {
	panic("impl")
}

func (ck *Clerk) BecomePrimary(args *BecomePrimaryArgs) Error {
	panic("impl")
}
