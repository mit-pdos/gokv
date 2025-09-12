package comulti

import (
	"github.com/mit-pdos/gokv/paxi/comulti/preparereply_gk"
	"github.com/mit-pdos/gokv/paxi/comulti/proposeargs_gk"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type Clerk struct {
	cl *urpc.Client
}

func MakeClerk(host uint64) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) Prepare(pn uint64, reply *preparereply_gk.S) {
	rawRep := new([]byte)
	ck.cl.Call(PREPARE, marshal.WriteInt(make([]byte, 8), pn), rawRep, 100 /* ms */)
	rep, _ := preparereply_gk.Unmarshal(*rawRep)
	*reply = rep
}

func (ck *Clerk) Propose(Pn uint64, CommitIndex uint64, Log []Entry) bool {
	rawRep := new([]byte)
	args := &proposeargs_gk.S{Pn: Pn, CommitIndex: CommitIndex, Log: Log}
	ck.cl.Call(PROPOSE, proposeargs_gk.Marshal(make([]byte, 0), *args), rawRep, 100 /* ms */)
	ret, _ := marshal.ReadBool(*rawRep)
	return ret
}
