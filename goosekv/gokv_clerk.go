package goosekv

import (
	"github.com/mit-pdos/lockservice/grove_ffi"
	"sync"
)

type GoKVClerk struct {
	mu  *sync.Mutex
	seq uint64
	cid uint64
	cl  *grove_ffi.RPCClient
}

func MakeKVClerk(cid uint64, host string) *GoKVClerk {
	ck := new(GoKVClerk)
	ck.cl = grove_ffi.MakeRPCClient(host)
	ck.cid = cid
	ck.seq = 1
	return ck
}

func MakeKVClerkWithRPCClient(cid uint64, cl *grove_ffi.RPCClient) *GoKVClerk {
	ck := new(GoKVClerk)
	ck.cl = cl
	ck.cid = cid
	ck.seq = 1
	return ck
}

func (ck *GoKVClerk) Put(key uint64, value []byte) {
	args := PutRequest{ck.cid, ck.seq, key, value}
	rawRep := make([]byte, 0)
	ck.cl.RemoteProcedureCall(KV_PUT, encodePutRequest(&args), &rawRep)
}
