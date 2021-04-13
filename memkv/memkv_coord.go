package memkv

import (
	"github.com/upamanyus/urpc/rpc"
	"sync"
)

const COORD_MOVE = 1
const COORD_GET = 2

type MemKVCoord struct {
	mu       *sync.Mutex
	seq      uint64
	cid      uint64
	cl       *rpc.RPCClient
	shardMap [NSHARD]HostName // maps from sid -> host that currently owns it
}

func (c *MemKVCoord) AddPeer() {
}

func (c *MemKVCoord) MoveShard(sid uint64, dst uint64) {
}

func (c *MemKVCoord) GetShardMapRPC(_ []byte, rep *[]byte) {
	c.mu.Lock()
	*rep = encodeShardMap(&c.shardMap)
	c.mu.Unlock()
}
