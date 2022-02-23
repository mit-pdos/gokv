package reconfig

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
)

type Clerk struct {
	cl *rpc.RPCClient
}

type OpType = []byte

type Error = uint64

const (
	ENone       = uint64(0)
	ENotPrimary = uint64(1)
	EStale      = uint64(2)
)

type DoOperationArgs struct {
	cn  uint64
	op  []byte
	osn uint64
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

func (ck *Clerk) DoOperation(args *DoOperationArgs) Error {
	// FIXME: impl
	return EStale
}

func (ck *Clerk) BecomeReplica(args *BecomeReplicaArgs) Error {
	// FIXME: impl
	return EStale
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	// FIXME: impl
	return nil
}
