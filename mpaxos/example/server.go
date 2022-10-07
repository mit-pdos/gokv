package example

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/mpaxos"
)

// if op is non-empty, set the state to that, otherwise return the current value
func applyFn(state []byte, op []byte) ([]byte, []byte) {
	if len(op) > 0 {
		return op[1:], make([]byte, 0)
	} else {
		return state, state
	}
}

type Clerk struct {
	c *mpaxos.Clerk
}

func MakeClerk(config []grove_ffi.Address) *Clerk {
	return &Clerk{c: mpaxos.MakeClerk(config)}
}

func StartServer(fname string, me grove_ffi.Address, config []grove_ffi.Address) {
	mpaxos.StartServer(fname, me, applyFn, config)
}

func (ck *Clerk) Put(val []byte) {
	op := make([]byte, len(val)+1)
	ck.c.Apply(op)
}

func (ck *Clerk) Get() []byte {
	return ck.c.Apply(make([]byte, 0))
}
