package memkv

import (
	"github.com/upamanyus/urpc/rpc"
)

type MemKVShardClerk struct {
	seq uint64
	cid uint64
	cl  *rpc.RPCClient
}

func MakeFreshKVClerk(host string) *MemKVShardClerk {
	ck := new(MemKVShardClerk)
	ck.cl = rpc.MakeRPCClient(host)
	rawRep := make([]byte, 0)
	ck.cl.Call(KV_FRESHCID, make([]byte, 0), &rawRep)
	ck.cid = decodeCID(rawRep)
	ck.seq = 1

	return ck
}

func MakeKVClerk(cid uint64, host string) *MemKVShardClerk {
	ck := new(MemKVShardClerk)
	ck.cl = rpc.MakeRPCClient(host)
	ck.cid = cid
	ck.seq = 1
	return ck
}

func MakeKVClerkWithRPCClient(cid uint64, cl *rpc.RPCClient) *MemKVShardClerk {
	ck := new(MemKVShardClerk)
	ck.cl = cl
	ck.cid = cid
	ck.seq = 1
	return ck
}

func (ck *MemKVShardClerk) Put(key uint64, value []byte) ErrorType {
	args := PutRequest{ck.cid, ck.seq, key, value}
	ck.seq = ck.seq + 1

	rawRep := make([]byte, 0)
	// TODO: helper for looping RemoteProcedureCall()
	for ck.cl.Call(KV_PUT, encodePutRequest(&args), &rawRep) == true {
	}
	return decodePutReply(rawRep).Err
}

func (ck *MemKVShardClerk) Get(key uint64, err *ErrorType, value *[]byte) {
	args := GetRequest{ck.cid, ck.seq, key}
	ck.seq = ck.seq + 1

	rawRep := make([]byte, 0)
	for ck.cl.Call(KV_GET, encodeGetRequest(&args), &rawRep) == true {
	}
	rep := decodeGetReply(rawRep)
	*err = rep.Err
	*value = rep.Value
}

func (ck *MemKVShardClerk) InstallShard(sid uint64, kvs map[uint64][]byte) {
	args := InstallShardRequest{CID:ck.cid, Seq:ck.seq, Sid:sid, Kvs:kvs}
	ck.seq = ck.seq + 1

	rawRep := make([]byte, 0)
	for ck.cl.Call(KV_INS_SHARD, encodeInstallShardRequest(&args), &rawRep) == true {
	}
}

func (ck *MemKVShardClerk) MoveShard(sid uint64, dst string) {
	args := MoveShardRequest{Sid:sid, Dst:dst}

	rawRep := make([]byte, 0)
	for ck.cl.Call(KV_MOV_SHARD, encodeMoveShardRequest(&args), &rawRep) == true {
	}
}
