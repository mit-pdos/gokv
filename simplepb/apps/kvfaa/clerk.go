package kvfaa

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/clerk"
	"github.com/tchajed/marshal"
)

type Clerk struct {
	cl *clerk.Clerk
}

func MakeClerk(configHost grove_ffi.Address) *Clerk {
	return &Clerk{cl: clerk.Make(configHost)}
}

func (ck *Clerk) FetchAndAppend(key uint64, val []byte) []byte {
	var args = make([]byte, 0, 8+len(val))
	args = marshal.WriteInt(args, key)
	args = marshal.WriteBytes(args, val)

	return ck.cl.Apply(args)
}
