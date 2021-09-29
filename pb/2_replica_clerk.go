package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
)

const REPLICA_APPEND = uint64(0)
const REPLICA_GETLOG = uint64(1)
const PRIMARY_ADDREPLICA = uint64(2)

type AppendArgs struct {
	cn        uint64
	log       []LogEntry
	commitIdx uint64
}

type ReplicaClerk struct {
	cl *rpc.RPCClient
}

func (ck *ReplicaClerk) AppendRPC(args AppendArgs) bool {
	// FIXME impl
	return false
}

// func (ck *ReplicaClerk) GetLogRPC()
