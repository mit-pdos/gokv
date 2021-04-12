package memkv

import (
	"github.com/upamanyus/urpc/rpc"
)

const COORD_MOVE = 1

type MemKVCoord struct {
	seq uint64
	cid uint64
	cl  *rpc.RPCClient
	shardMap [NSHARD]string // maps from sid -> host that currently owns it
}

func (c *MemKVCoord) MoveShard(sid uint64, dst uint64) {
}
