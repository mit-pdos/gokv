package replica

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

type Error = uint64

const (
	ENone             = uint64(0)
	ENotPrimary       = uint64(1)
	EStale            = uint64(2)
	EAppendOutOfOrder = uint64(3)
	ETruncated        = uint64(4)
	EIncompleteLog    = uint64(5)
)

func (ck *Clerk) appendRPC(args *AppendArgs) Error {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) BecomePrimary(args *BecomePrimaryArgs) Error {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) TryBecomeReplica(args *BecomeReplicaArgs) Error {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) RemainReplica(args *BecomeReplicaArgs) Error {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) GetUncommittedLog(epoch uint64) (Error, uint64, []LogEntry) {
	panic("replica: impl")
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}
