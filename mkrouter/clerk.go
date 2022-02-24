package mkrouter

import (
	"github.com/mit-pdos/gokv/memkv"
	"github.com/mit-pdos/gokv/urpc/rpc"
)

type Clerk struct {
	cl *rpc.RPCClient
}

func (ck *Clerk) Get(key uint64) []byte {
	var ret []byte
	ck.cl.Call(memkv.KV_GET, memkv.EncodeGetRequest(&memkv.GetRequest{Key: key}), &ret, 100 /*ms*/)
	return ret
}

func (ck *Clerk) Put(key uint64, value []byte) {
	var ret []byte
	ck.cl.Call(memkv.KV_GET, memkv.EncodePutRequest(&memkv.PutRequest{Key: key, Value: value}), &ret, 100 /*ms*/)
}
