package memkv

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
)

type MemKVShardClerk struct {
	seq uint64
	cid uint64
	cl  *rpc.RPCClient
}

func MakeFreshKVClerk(host HostName) *MemKVShardClerk {
	ck := new(MemKVShardClerk)
	ck.cl = rpc.MakeRPCClient(host)
	rawRep := new([]byte)
	for ck.cl.Call(KV_FRESHCID, make([]byte, 0), rawRep) == true {
	}
	ck.cid = decodeCID(*rawRep)
	ck.seq = 1

	return ck
}

func (ck *MemKVShardClerk) Put(key uint64, value []byte) ErrorType {
	args := new(PutRequest)
	args.CID = ck.cid
	args.Seq = ck.seq
	args.Key = key
	args.Value = value
	ck.seq = ck.seq + 1

	rawRep := new([]byte)
	// TODO: helper for looping RemoteProcedureCall()
	for ck.cl.Call(KV_PUT, encodePutRequest(args), rawRep) == true {
	}
	return decodePutReply(*rawRep).Err
}

func (ck *MemKVShardClerk) Get(key uint64, value *[]byte) ErrorType {
	args := new(GetRequest)
	args.CID = ck.cid
	args.Seq = ck.seq
	args.Key = key
	ck.seq = ck.seq + 1

	rawRep := new([]byte)
	for ck.cl.Call(KV_GET, encodeGetRequest(args), rawRep) == true {
	}
	rep := decodeGetReply(*rawRep)
	*value = rep.Value
	return rep.Err
}

func (ck *MemKVShardClerk) InstallShard(sid uint64, kvs map[uint64][]byte) {
	args := new(InstallShardRequest)
	args.CID = ck.cid
	args.Seq = ck.seq
	args.Sid = sid
	args.Kvs = kvs
	ck.seq = ck.seq + 1

	rawRep := new([]byte)
	for ck.cl.Call(KV_INS_SHARD, encodeInstallShardRequest(args), rawRep) == true {
	}
}

func (ck *MemKVShardClerk) MoveShard(sid uint64, dst HostName) {
	args := MoveShardRequest{Sid: sid, Dst: dst}

	rawRep := make([]byte, 0)
	for ck.cl.Call(KV_MOV_SHARD, encodeMoveShardRequest(&args), &rawRep) == true {
	}
}
