package memkv

import (
	"github.com/mit-pdos/gokv/connman"
)

type MemKVCoordClerk struct {
	host HostName
	c *connman.ConnMan
}

func (ck *MemKVCoordClerk) AddShardServer(dst HostName) {
	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, COORD_ADD, EncodeUint64(dst), rawRep, 10000 /*ms*/)
	return
}

func (ck *MemKVCoordClerk) GetShardMap() []HostName {
	rawRep := new([]byte)
	ck.c.CallAtLeastOnce(ck.host, COORD_GET, make([]byte, 0), rawRep, 100 /*ms*/)
	return decodeShardMap(*rawRep)
}

// NOTE: a single clerk keeps quite a bit of state, via the shardMap[], so it
// might be good to not need to duplicate shardMap[] for a pool of clerks that's
// safe for concurrent use
type MemKVClerk struct {
	shardClerks *ShardClerkSet
	coordCk     *MemKVCoordClerk
	shardMap    []HostName // size == NSHARD; maps from sid -> host that currently owns it
}

func (ck *MemKVClerk) Get(key uint64) []byte {
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

func (ck *MemKVClerk) Put(key uint64, value []byte) {
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

func (ck *MemKVClerk) ConditionalPut(key uint64, expectedValue []byte, newValue []byte) bool {
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

func (ck *MemKVClerk) Add(host HostName) {
	ck.coordCk.AddShardServer(host)
}

func MakeMemKVClerk(coord HostName) *MemKVClerk {
	c := connman.MakeConnMan()
	cck := new(MemKVCoordClerk)
	ck := new(MemKVClerk)
	ck.coordCk = cck
	ck.coordCk.host = coord
	ck.coordCk.c = c
	ck.shardClerks = MakeShardClerkSet(c)
	ck.shardMap = ck.coordCk.GetShardMap()
	return ck
}

// TODO: add an Append(key, value) (oldValue []byte) call
