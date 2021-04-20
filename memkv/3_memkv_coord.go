package memkv

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
)

const COORD_MOVE = uint64(1)
const COORD_GET = uint64(2)

type ShardClerkSet struct {
	cls map[HostName]*MemKVShardClerk
}

func MakeShardClerkSet() *ShardClerkSet {
	return &ShardClerkSet{cls: make(map[HostName]*MemKVShardClerk)}
}

func (s *ShardClerkSet) GetClerk(host HostName) *MemKVShardClerk {
	ck, ok := s.cls[host]
	if !ok {
		ck2 := MakeFreshKVClerk(host)
		s.cls[host] = ck2
		return ck2
	} else {
		return ck
	}
}

type MemKVCoord struct {
	mu          *sync.Mutex
	config      map[HostName]string
	shardMap    []HostName          // maps from sid -> host that currently owns it
	hostShards  map[HostName]uint64 // maps from host -> num shard that it currently has
	shardClerks ShardClerkSet
}

func (c *MemKVCoord) AddServerRPC(newhost HostName) {
	c.mu.Lock()
	// Greedily rebalances shards using minimum number of migrations

	// currently, (NSHARD/numHosts) +/- 1 shard should be assigned to each server
	//
	// We keep a map[HostName]uint64 to remember how many shards we've given
	// each shard server. Then, we iterate over shardMap[], and move a shard if the current holder does.
	//
	// (NSHARD - numHosts * floor(NSHARD/numHosts)) will have size (floor(NSHARD/numHosts) + 1)
	numHosts := uint64(10)
	numShardFloor := NSHARD / numHosts
	numShardCeil := NSHARD/numHosts + 1
	nf_left := numHosts - (NSHARD - numHosts*NSHARD/numHosts) // number of servers that will have one fewer shard than other servers
	for sid, host := range c.shardMap {
		n := c.hostShards[host]
		if n > numShardFloor {
			if n == numShardCeil {
				if nf_left > 0 {
					nf_left = nf_left - 1
					c.hostShards[host] = n - 1
					c.shardClerks.GetClerk(host).MoveShard(uint64(sid), newhost)
				}
				// else, we have already made enough hosts have the minimum number of shard servers
			} else {
				c.hostShards[host] = n - 1
				c.shardClerks.GetClerk(host).MoveShard(uint64(sid), newhost)
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

func MakeMemKVCoordServer(initserver string) *MemKVCoord {
	s := new(MemKVCoord)
	s.mu = new(sync.Mutex)

	s.shardMap = make([]HostName, NSHARD)
	for i := uint64(0); i < NSHARD; i++ {
		s.shardMap[i] = initserver
	}
	return s
}

func (c *MemKVCoord) Start(host string) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[COORD_GET] = c.GetShardMapRPC
	s := rpc.MakeRPCServer(handlers)
	s.Serve(host, 1)
}
