package example

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

func MakeClerk(grove_ffi.Address) Clerk {
	panic("example: impl")
}

func (ck *Clerk) GetStateAndStopTruncation() (uint64, []byte) {
	panic("example: impl")
}

func (ck *Clerk) SetState(index uint64, state []byte) {
	panic("example: impl")
}
