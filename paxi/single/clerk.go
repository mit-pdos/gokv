package single

import (
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

func MakeClerk(host uint64) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) Prepare(pn uint64, reply *PrepareReply) {
	// pass
}

func (ck *Clerk) Propose(Pn uint64, Val ValType) bool {
	// pass
	return false
}
