package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
)

type ReplicaClerk struct {
	cl *rpc.RPCClient
}

func (ck *ReplicaClerk) AppendRPC(args AppendArgs) bool {
	// FIXME impl
	return false
}

// func (ck *ReplicaClerk) GetLogRPC()
