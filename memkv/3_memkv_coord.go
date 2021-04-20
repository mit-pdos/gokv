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
	hostShards map[HostName]uint64 // maps from host -> num shard that it currently has
}

func (c *MemKVCoord) AddServerRPC(host string) {
	c.mu.Lock()
	// Greedily rebalances shards using minimum number of migrations

	// currently, (NSHARD/numHosts) +/- 1 shard should be assigned to each server
	//
	// We keep a map[HostName]uint64 to remember how many shards we've given
	// each shard server. Then, we iterate over shardMap[], and move a shard if the current holder does.
	//
	// (NSHARD - numHosts * floor(NSHARD/numHosts)) will have size (floor(NSHARD/numHosts) + 1)
	numHosts := uint64(10)
	numShardFloor := NSHARD/numHosts
	numShardCeil := NSHARD/numHosts + 1
	nf_left := numHosts - (NSHARD - numHosts * NSHARD/numHosts) // number of servers that will have one fewer shard than other servers
	for _, host := range c.shardMap {
		n := c.hostShards[host]
		if n > numShardFloor {
			if n == numShardCeil {
				if nf_left > 0 {
					nf_left = nf_left - 1
					c.hostShards[host] = n - 1
					// FIXME: MoveShardRPC()
				}
				// else, we have already made enough hosts have the minimum number of shard servers
			} else {
				c.hostShards[host] = n - 1
				// FIXME: MoveShardRPC()
			}
		}
	}
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
