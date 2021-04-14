package memkv

import (
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"sync"
)

const COORD_MOVE = uint64(1)
const COORD_GET = uint64(2)

type MemKVCoord struct {
	mu       *sync.Mutex
	config   map[HostName]string
	shardMap []HostName // maps from sid -> host that currently owns it
}

func (c *MemKVCoord) AddServerRPC(host string) {
	c.mu.Lock()
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
	s.config = make(map[HostName]string)
	s.config[1] = "localhost:37001"
	s.config[2] = "localhost:37002"

	for i := uint64(0); i < NSHARD; i++ {
		s.shardMap[i] = i % 2 // s.config[i%uint64(len(s.shardServers))]
	}
	return s
}

func (c *MemKVCoord) Start() {
	handlers := make(map[uint64]grove_common.RawRpcFunc)
	handlers[COORD_GET] = c.GetShardMapRPC
	grove_ffi.StartRPCServer(handlers)
}
