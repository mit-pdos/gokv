package memkv

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
)

type MemKVCoordClerk struct {
	cl *rpc.RPCClient
}

func (ck *MemKVCoordClerk) AddShardServer(dst string) {
	rawRep := new([]byte)
	for ck.cl.Call(COORD_ADD, []byte(dst), rawRep) == true {
	}
	return
}

func (ck *MemKVCoordClerk) GetShardMap() []HostName {
	rawRep := new([]byte)
	for ck.cl.Call(COORD_GET, make([]byte, 0), rawRep) == true {
	}
	return decodeShardMap(*rawRep)
}

// NOTE: a single clerk keeps quite a bit of state, via the shardMap[], so it
// might be good to not need to duplicate shardMap[] for a pool of clerks that's
// safe for concurrent use
type MemKVClerk struct {
	shardClerks *ShardClerkSet
	coordCk     MemKVCoordClerk
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

func (ck *MemKVClerk) Add(host string) {
	ck.coordCk.AddShardServer(host)
}

func MakeMemKVClerk(coord HostName) *MemKVClerk {
	ck := new(MemKVClerk)
	ck.coordCk.cl = rpc.MakeRPCClient(coord)
	ck.shardClerks = MakeShardClerkSet()
	ck.shardMap = ck.coordCk.GetShardMap()
	return ck
}

// TODO: add an Append(key, value) (oldValue []byte) call
