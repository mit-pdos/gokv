package reconfig

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

type OpType = []byte

type Error = uint64

const (
	ENone             = uint64(0)
	ENotPrimary       = uint64(1)
	EStale            = uint64(2)
	EAppendOutOfOrder = uint64(3)
)

func (ck *Clerk) appendRPC(args *AppendArgs) Error {
	// FIXME: impl
	return EStale
}

func (ck *Clerk) BecomeReplica(args *BecomeReplicaArgs) Error {
	// FIXME: impl
	return EStale
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}
