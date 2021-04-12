package memkv

import (
	"github.com/upamanyus/urpc/rpc"
)

type MemKVCoordClerk struct {
	seq uint64
	cid uint64
	cl  *rpc.RPCClient
	shardMap [NSHARD]string // maps from sid -> host that currently owns it
}

func (ck *MemKVCoordClerk) MoveShard(sid uint64, dst uint64) {
}
