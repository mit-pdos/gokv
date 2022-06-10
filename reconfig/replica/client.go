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

type AppendArgs struct {
	cn    uint64
	entry LogEntry
	index uint64
}

type Configuration struct {
	replicas []grove_ffi.Address
}

type BecomeReplicaArgs struct {
	cn    uint64
	state []byte
	osn   uint64
}

type BecomePrimaryArgs struct {
	conf    Configuration
	repArgs *BecomeReplicaArgs
}

type GetStateReply struct {
	cn    uint64
	state []byte
	osn   uint64
}

func (ck *Clerk) DoOperation(args *AppendArgs) Error {
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
