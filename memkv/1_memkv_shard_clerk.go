package memkv

import (
	"github.com/mit-pdos/gokv/connman"
	"github.com/goose-lang/std"
)

type MemKVShardClerk struct {
	seq uint64
	cid uint64
	host HostName
	c *connman.ConnMan
}

func MakeFreshKVClerk(host HostName) *MemKVShardClerk {
	ck := new(MemKVShardClerk)
	ck.host = host
	ck.c = connman.MakeConnMan()
	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(host, KV_FRESHCID, make([]byte, 0), rawRep, 100/*ms*/)
	ck.cid = DecodeUint64(*rawRep)
	ck.seq = 1

	return ck
}

func (ck *MemKVShardClerk) Put(key uint64, value []byte) ErrorType {
	args := new(PutRequest)
	args.CID = ck.cid
	args.Seq = ck.seq
	args.Key = key
	args.Value = value
	// Overflowing a 64bit counter will take a while, assume it dos not happen
	ck.seq = std.SumAssumeNoOverflow(ck.seq, 1)

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_PUT, EncodePutRequest(args), rawRep, 100/*ms*/)
	rep := DecodePutReply(*rawRep)
	return rep.Err
}

func (ck *MemKVShardClerk) Get(key uint64, value *[]byte) ErrorType {
	args := new(GetRequest)
	args.CID = ck.cid
	args.Seq = ck.seq
	args.Key = key
	// Overflowing a 64bit counter will take a while, assume it dos not happen
	ck.seq = std.SumAssumeNoOverflow(ck.seq, 1)

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_GET, EncodeGetRequest(args), rawRep, 100/*ms*/)
	rep := DecodeGetReply(*rawRep)
	*value = rep.Value
	return rep.Err
}

func (ck *MemKVShardClerk) ConditionalPut(key uint64, expectedValue []byte, newValue []byte, success *bool) ErrorType {
	args := new(ConditionalPutRequest)
	args.CID = ck.cid
	args.Seq = ck.seq
	args.Key = key
	args.ExpectedValue = expectedValue
	args.NewValue = newValue
	// Overflowing a 64bit counter will take a while, assume it dos not happen
	ck.seq = std.SumAssumeNoOverflow(ck.seq, 1)

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_CONDITIONAL_PUT, EncodeConditionalPutRequest(args), rawRep, 100/*ms*/)
	rep := DecodeConditionalPutReply(*rawRep)
	*success = rep.Success
	return rep.Err
}

func (ck *MemKVShardClerk) InstallShard(sid uint64, kvs map[uint64][]byte) {
	// log.Printf("InstallShard %d starting", sid)
	args := new(InstallShardRequest)
	args.CID = ck.cid
	args.Seq = ck.seq
	args.Sid = sid
	args.Kvs = kvs
	// Overflowing a 64bit counter will take a while, assume it dos not happen
	ck.seq = std.SumAssumeNoOverflow(ck.seq, 1)

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_INS_SHARD, encodeInstallShardRequest(args), rawRep, 100/*ms*/)
	// log.Printf("InstallShard %d finished", sid)
}

func (ck *MemKVShardClerk) MoveShard(sid uint64, dst HostName) {
	args := new(MoveShardRequest)
	args.Sid = sid
	args.Dst = dst

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_MOV_SHARD, encodeMoveShardRequest(args), rawRep, 100/*ms*/)
}
