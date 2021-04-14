package memkv

import (
	"github.com/upamanyus/urpc/rpc"
)

type MemKVCoordClerk struct {
	seq      uint64
	cid      uint64
	cl       *rpc.RPCClient
	shardMap [NSHARD]HostName // maps from sid -> host that currently owns it
}

func (ck *MemKVCoordClerk) MoveShard(sid uint64, dst uint64) {
}

func (ck *MemKVCoordClerk) GetShardMap() []HostName {
	rawRep := new([]byte)
	ck.cl.Call(COORD_GET, make([]byte, 0), rawRep)
	return decodeShardMap(*rawRep)
}

type ShardClerkSet struct {
	cls map[HostName]*MemKVShardClerk
}

func (s *ShardClerkSet) getClerk(host HostName) *MemKVShardClerk {
	ck, ok := s.cls[host]
	if !ok {
		ck2 := MakeFreshKVClerk(host)
		s.cls[host] = ck2
		return ck2
	} else {
		return ck
	}
}

// NOTE: a single clerk keeps quite a bit of state, via the shardMap[], so it
// might be good to not need to duplicate shardMap[] for a pool of clerks that's
// safe for concurrent use
type MemKVClerk struct {
	seq         uint64
	cid         uint64
	shardClerks *ShardClerkSet
	coordCk     MemKVCoordClerk
	shardMap    []HostName // size == NSHARD; maps from sid -> host that currently owns it
}

func (ck *MemKVClerk) Get(key uint64) []byte {
	val := new([]byte)
	for {
		sid := shardOf(key)
		shardServer := ck.shardMap[sid]

		shardCk := ck.shardClerks.getClerk(shardServer)
		err := shardCk.Get(key, val)
		if err == EDontHaveShard {
			ck.shardMap = ck.coordCk.GetShardMap()
		} else if err == ENone {
			break
		}
	}
	return *val
}

func (ck *MemKVClerk) Put(key uint64, value []byte) {
	for {
		sid := shardOf(key)
		shardServer := ck.shardMap[sid]

		shardCk := ck.shardClerks.getClerk(shardServer)
		err := shardCk.Put(key, value)

		if err == EDontHaveShard {
			ck.shardMap = ck.coordCk.GetShardMap()
		} else if err == ENone {
			break
		}
	}
	return
}

// TODO: add an Append(key, value) (oldValue []byte) call
