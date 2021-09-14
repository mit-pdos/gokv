package memkv

import (
	"github.com/mit-pdos/gokv/connman"
)

type KVCoordClerk struct {
	host HostName
	c    *connman.ConnMan
}

func (ck *KVCoordClerk) AddShardServer(dst HostName) {
	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, COORD_ADD, EncodeUint64(dst), rawRep, 50000 /*ms*/)
	return
}

func (ck *KVCoordClerk) GetShardMap() []HostName {
	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, COORD_GET, make([]byte, 0), rawRep, 2000 /*ms*/)
	return decodeShardMap(*rawRep)
}

// "Sequential" KV clerk, can only be used for one request at a time.
// NOTE: a single clerk keeps quite a bit of state, via the shardMap[], so it
// might be good to not need to duplicate shardMap[] for a pool of clerks that's
// safe for concurrent use
type SeqKVClerk struct {
	shardClerks *ShardClerkSet
	coordCk     *KVCoordClerk
	shardMap    []HostName // size == NSHARD; maps from sid -> host that currently owns it
}

func (ck *SeqKVClerk) Get(key uint64) []byte {
	val := new([]byte)
	for {
		sid := shardOf(key)
		shardServer := ck.shardMap[sid]

		shardCk := ck.shardClerks.GetClerk(shardServer)
		err := shardCk.Get(key, val)
		if err == ENone {
			break
		}
		ck.shardMap = ck.coordCk.GetShardMap()
		continue
	}
	return *val
}

func (ck *SeqKVClerk) Put(key uint64, value []byte) {
	for {
		sid := shardOf(key)
		shardServer := ck.shardMap[sid]

		shardCk := ck.shardClerks.GetClerk(shardServer)
		err := shardCk.Put(key, value)

		if err == ENone {
			break
		}
		ck.shardMap = ck.coordCk.GetShardMap()
		continue
	}
	return
}

func (ck *SeqKVClerk) ConditionalPut(key uint64, expectedValue []byte, newValue []byte) bool {
	success := new(bool)
	for {
		sid := shardOf(key)
		shardServer := ck.shardMap[sid]

		shardCk := ck.shardClerks.GetClerk(shardServer)
		err := shardCk.ConditionalPut(key, expectedValue, newValue, success)

		if err == ENone {
			break
		}
		ck.shardMap = ck.coordCk.GetShardMap()
		continue
	}
	return *success
}

func (ck *SeqKVClerk) Add(host HostName) {
	ck.coordCk.AddShardServer(host)
}

func MakeSeqKVClerk(coord HostName, cm *connman.ConnMan) *SeqKVClerk {
	cck := new(KVCoordClerk)
	ck := new(SeqKVClerk)
	ck.coordCk = cck
	ck.coordCk.host = coord
	ck.coordCk.c = cm
	ck.shardClerks = MakeShardClerkSet(cm)
	ck.shardMap = ck.coordCk.GetShardMap()
	return ck
}

// TODO: add an Append(key, value) (oldValue []byte) call
