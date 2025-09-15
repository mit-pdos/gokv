package mkrouter

import (
	"github.com/mit-pdos/gokv/memkv/getrequest_gk"
	"github.com/mit-pdos/gokv/memkv/kvop_gk"
	"github.com/mit-pdos/gokv/memkv/putrequest_gk"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

func (ck *Clerk) Get(key uint64) []byte {
	var ret []byte
	ck.cl.Call(uint64(kvop_gk.KV_GET), getrequest_gk.Marshal(make([]byte, 0), getrequest_gk.S{Key: key}), &ret, 100 /*ms*/)
	return ret
}

func (ck *Clerk) Put(key uint64, value []byte) {
	var ret []byte
	ck.cl.Call(uint64(kvop_gk.KV_PUT), putrequest_gk.Marshal(make([]byte, 0), putrequest_gk.S{Key: key, Value: value}), &ret, 100 /*ms*/)
}
