package memkv

import (
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/erpc"
	"github.com/mit-pdos/gokv/memkv/conditionalputreply_gk"
	"github.com/mit-pdos/gokv/memkv/conditionalputrequest_gk"
	"github.com/mit-pdos/gokv/memkv/error_gk"
	"github.com/mit-pdos/gokv/memkv/getreply_gk"
	"github.com/mit-pdos/gokv/memkv/getrequest_gk"
	"github.com/mit-pdos/gokv/memkv/kvop_gk"
	"github.com/mit-pdos/gokv/memkv/moveshardrequest_gk"
	"github.com/mit-pdos/gokv/memkv/putreply_gk"
	"github.com/mit-pdos/gokv/memkv/putrequest_gk"
	"github.com/tchajed/marshal"
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
	ck.c.CallAtLeastOnce(host, uint64(kvop_gk.KV_FRESHCID), make([]byte, 0), rawRep, 100 /*ms*/)
	cid, _ := marshal.ReadInt(*rawRep)
	ck.erpc = erpc.MakeClient(cid)

	return ck
}

func (ck *KVShardClerk) Put(key uint64, value []byte) error_gk.E {
	args := new(putrequest_gk.S)
	args.Key = key
	args.Value = value
	req := ck.erpc.NewRequest(putrequest_gk.Marshal(make([]byte, 0), *args))

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, uint64(kvop_gk.KV_PUT), req, rawRep, 100 /*ms*/)
	rep, _ := putreply_gk.Unmarshal(*rawRep)
	return rep.Err
}

func (ck *KVShardClerk) Get(key uint64, value *[]byte) error_gk.E {
	args := new(getrequest_gk.S)
	args.Key = key
	req := ck.erpc.NewRequest(getrequest_gk.Marshal(make([]byte, 0), *args))

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, uint64(kvop_gk.KV_GET), req, rawRep, 100 /*ms*/)
	rep, _ := getreply_gk.Unmarshal(*rawRep)
	*value = rep.Value
	return rep.Err
}

func (ck *KVShardClerk) ConditionalPut(key uint64, expectedValue []byte, newValue []byte, success *bool) error_gk.E {
	args := new(conditionalputrequest_gk.S)
	args.Key = key
	args.ExpectedValue = expectedValue
	args.NewValue = newValue
	req := ck.erpc.NewRequest(conditionalputrequest_gk.Marshal(make([]byte, 0), *args))

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, uint64(kvop_gk.KV_CONDITIONAL_PUT), req, rawRep, 100 /*ms*/)
	rep, _ := conditionalputreply_gk.Unmarshal(*rawRep)
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
	ck.c.CallAtLeastOnce(ck.host, uint64(kvop_gk.KV_INS_SHARD), req, rawRep, 100 /*ms*/)
	// log.Printf("InstallShard %d finished", sid)
}

func (ck *KVShardClerk) MoveShard(sid uint64, dst HostName) {
	args := new(moveshardrequest_gk.S)
	args.Sid = sid
	args.Dst = dst

	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, uint64(kvop_gk.KV_MOV_SHARD), moveshardrequest_gk.Marshal(make([]byte, 0), *args), rawRep, 100 /*ms*/)
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
