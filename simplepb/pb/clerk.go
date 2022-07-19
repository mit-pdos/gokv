package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	ck := &Clerk{cl: urpc.MakeClient(host)}
	return ck
}

func (ck *Clerk) Apply(epoch uint64, index uint64, op Op) Error {
	panic("impl")
}
