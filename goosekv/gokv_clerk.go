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

func (ck *GoKVClerk) Put(key uint64, value []byte) ErrorType {
	args := PutRequest{ck.cid, ck.seq, key, value}
	ck.seq = ck.seq + 1

	rawRep := make([]byte, 0)
	// TODO: helper for looping RemoteProcedureCall()
	for ck.cl.RemoteProcedureCall(KV_PUT, encodePutRequest(&args), &rawRep) == true {
	}
	return decodePutReply(rawRep).Err
}

func (ck *GoKVClerk) Get(key uint64, err *ErrorType, value *[]byte) {
	args := GetRequest{ck.cid, ck.seq, key}
	ck.seq = ck.seq + 1

	rawRep := make([]byte, 0)
	for ck.cl.RemoteProcedureCall(KV_GET, encodeGetRequest(&args), &rawRep) == true {
	}
	rep := decodeGetReply(rawRep)
	*err = rep.Err
	*value = rep.Value
}
