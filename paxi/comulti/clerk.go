package comulti

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
	rawRep := new([]byte)
	ck.cl.Call(PREPARE, encodeUint64(pn), rawRep, 100 /* ms */)
	*reply = *(decodePrepareReply(*rawRep))
}

func (ck *Clerk) Propose(Pn uint64, CommitIndex uint64, Log []Entry) bool {
	rawRep := new([]byte)
	args := &ProposeArgs{Pn: Pn, CommitIndex: CommitIndex, Log: Log}
	ck.cl.Call(PROPOSE, encodeProposeArgs(args), rawRep, 100 /* ms */)
	return decodeBool(*rawRep)
}
