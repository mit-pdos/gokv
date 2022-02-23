package reconfig

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
)

type Clerk struct {
	cl *rpc.RPCClient
}

type DoOperationArgs struct {
	cn uint64
	op []byte
	osn uint64
}

type Configuration struct {
	replicas []grove_ffi.Address
}

type BecomeReplicaArgs struct {
	cn uint64
	state []byte
	osn uint64
}

type BecomePrimaryArgs struct {
	conf Configuration
	repArgs *BecomeReplicaArgs
}

func (ck *Clerk) DoOperation(args *DoOperationArgs) bool {
	return false
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	// FIXME: impl
	return nil
}
