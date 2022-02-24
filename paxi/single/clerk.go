package single

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
)

type Clerk struct {
	cl *rpc.RPCClient
}

func MakeClerk(host uint64) *Clerk {
	return &Clerk{cl: rpc.MakeRPCClient(host)}
}

func (ck *Clerk) Prepare(pn uint64, reply *PrepareReply) {
	// pass
}

func (ck *Clerk) Propose(Pn uint64, Val ValType) bool {
	// pass
	return false
}
