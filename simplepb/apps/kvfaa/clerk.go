package state

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/marshal"
)

type Clerk struct {
	cl *pb.Clerk
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: pb.MakeClerk(host)}
}

func (ck *Clerk) FetchAndAppend(key uint64, val []byte) (e.Error, []byte) {
	var args = make([]byte, 0, 8+len(val))
	args = marshal.WriteInt(args, key)
	args = marshal.WriteBytes(args, val)

	return ck.cl.Apply(args)
}
