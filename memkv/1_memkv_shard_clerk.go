package memkv

import (
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/erpc"
)

type KVShardClerk struct {
	erpc *erpc.Client
	host HostName
	c    *connman.ConnMan
}

func MakeFreshKVShardClerk(host HostName, c *connman.ConnMan) *KVShardClerk {
	ck := new(KVShardClerk)
	ck.host = host
	ck.c = c
	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(host, KV_FRESHCID, make([]byte, 0), rawRep, 100 /*ms*/)
	cid := DecodeUint64(*rawRep)
	ck.erpc = erpc.MakeClient(cid)

	return ck
}

func (ck *KVShardClerk) Put(key uint64, value []byte) ErrorType {
	args := new(PutRequest)
	args.Key = key
	args.Value = value
	req := ck.erpc.NewRequest(EncodePutRequest(args))

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_PUT, req, rawRep, 100 /*ms*/)
	rep := DecodePutReply(*rawRep)
	return rep.Err
}

func (ck *KVShardClerk) Get(key uint64, value *[]byte) ErrorType {
	args := new(GetRequest)
	args.Key = key
	req := ck.erpc.NewRequest(EncodeGetRequest(args))

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_GET, req, rawRep, 100 /*ms*/)
	rep := DecodeGetReply(*rawRep)
	*value = rep.Value
	return rep.Err
}

func (ck *KVShardClerk) ConditionalPut(key uint64, expectedValue []byte, newValue []byte, success *bool) ErrorType {
	args := new(ConditionalPutRequest)
	args.Key = key
	args.ExpectedValue = expectedValue
	args.NewValue = newValue
	req := ck.erpc.NewRequest(EncodeConditionalPutRequest(args))

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_CONDITIONAL_PUT, req, rawRep, 100 /*ms*/)
	rep := DecodeConditionalPutReply(*rawRep)
	*success = rep.Success
	return rep.Err
}

func (ck *KVShardClerk) InstallShard(sid uint64, kvs map[uint64][]byte) {
	// log.Printf("InstallShard %d starting", sid)
	args := new(InstallShardRequest)
	args.Sid = sid
	args.Kvs = kvs
	req := ck.erpc.NewRequest(encodeInstallShardRequest(args))

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_INS_SHARD, req, rawRep, 100 /*ms*/)
	// log.Printf("InstallShard %d finished", sid)
}

func (ck *KVShardClerk) MoveShard(sid uint64, dst HostName) {
	args := new(MoveShardRequest)
	args.Sid = sid
	args.Dst = dst

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, KV_MOV_SHARD, encodeMoveShardRequest(args), rawRep, 100 /*ms*/)
}

// The coordinator, and the main clerk, need to talk to a bunch of shards.
type ShardClerkSet struct {
	cls map[HostName]*KVShardClerk
	c   *connman.ConnMan
}

func MakeShardClerkSet(c *connman.ConnMan) *ShardClerkSet {
	return &ShardClerkSet{cls: make(map[HostName]*KVShardClerk), c: c}
}

func (s *ShardClerkSet) GetClerk(host HostName) *KVShardClerk {
	ck, ok := s.cls[host]
	if !ok {
		ck2 := MakeFreshKVShardClerk(host, s.c)
		s.cls[host] = ck2
		return ck2
	} else {
		return ck
	}
}
