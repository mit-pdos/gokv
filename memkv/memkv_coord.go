package memkv

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"sync"
)

const COORD_MOVE = 1
const COORD_GET = 2

type MemKVCoord struct {
	mu           *sync.Mutex
	shardServers []HostName
	shardMap     [NSHARD]HostName // maps from sid -> host that currently owns it
}

func (c *MemKVCoord) AddServerRPC(host string) {
	c.mu.Lock()
	c.shardServers = append(c.shardServers, host)
	panic("shard rebalancing unimpl")
	c.mu.Unlock()
}

func (c *MemKVCoord) GetShardMapRPC(_ []byte, rep *[]byte) {
	c.mu.Lock()
	*rep = encodeShardMap(&c.shardMap)
	c.mu.Unlock()
}

func MakeMemKVCoordServer() *MemKVCoord {
	s := new(MemKVCoord)
	s.mu = new(sync.Mutex)
	s.shardServers = []string{"localhost:37001", "localhost:37002"}
	for i := 0; i < NSHARD; i++ {
		s.shardMap[i] = s.shardServers[i%len(s.shardServers)]
	}
	return s
}

func (c *MemKVCoord) Start() {
	handlers := make(map[uint64]grove_common.RawRpcFunc)
	handlers[COORD_GET] = c.GetShardMapRPC
	grove_ffi.StartRPCServer(handlers)
}
