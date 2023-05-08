package kvee

import (
	"log"

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
	log.Print("Running PUT RW")
	ck.cl.ApplyExactlyOnce(kv64.EncodePutArgs(putArgs))
}

func (ck *Clerk) Get(key uint64) []byte {
	log.Print("Running GET RO")
	return ck.cl.ApplyReadonly(kv64.EncodeGetArgs(key))
}
