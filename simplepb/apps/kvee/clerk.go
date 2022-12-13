package kvee

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/eesm"
	"github.com/mit-pdos/gokv/simplepb/apps/kv64"
)

type Clerk struct {
	cl *eesm.Clerk
}

func MakeClerk(confHost grove_ffi.Address) *Clerk {
	return &Clerk{cl: eesm.MakeClerk(confHost)}
}

func (ck *Clerk) Put(key uint64, val []byte) {
	putArgs := &kv64.PutArgs{
		Key: key,
		Val: val,
	}
	ck.cl.ApplyExactlyOnce(kv64.EncodePutArgs(putArgs))
}

func (ck *Clerk) Get(key uint64) []byte {
	return ck.cl.ApplyExactlyOnce(kv64.EncodeGetArgs(key))
}
