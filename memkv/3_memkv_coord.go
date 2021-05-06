package memkv

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"log"
	"sync"
)

const COORD_ADD = uint64(1)
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
	shardClerks *ShardClerkSet
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
	log.Printf("Rebalancing\n")
	c.hostShards[newhost] = 0
	numHosts := uint64(len(c.hostShards))
	numShardFloor := NSHARD / numHosts
	numShardCeil := NSHARD/numHosts + 1
	var nf_left uint64
	nf_left = numHosts - (NSHARD - numHosts*NSHARD/numHosts) // number of servers that will have one fewer shard than other servers
	for sid, host := range c.shardMap {
		n := c.hostShards[host]
		if n > numShardFloor {
			if n == numShardCeil {
				if nf_left > 0 {
					nf_left = nf_left - 1
					// log.Printf("Moving %d from %s -> %s", sid, host, newhost)
					c.shardClerks.GetClerk(host).MoveShard(uint64(sid), newhost)
					c.hostShards[host] = n - 1
					c.hostShards[newhost] += 1
					c.shardMap[sid] = newhost
				}
				// else, we have already made enough hosts have the minimum number of shard servers
			} else {
				// log.Printf("Moving %d from %s -> %s", sid, host, newhost)
				c.shardClerks.GetClerk(host).MoveShard(uint64(sid), newhost)
				c.hostShards[host] = n - 1
				c.hostShards[newhost] += 1
				c.shardMap[sid] = newhost
			}
		}
	}
	log.Println("Done rebalancing")
	log.Printf("%+v", c.hostShards)
	c.mu.Unlock()
}

func (c *MemKVCoord) GetShardMapRPC(_ []byte, rep *[]byte) {
	c.mu.Lock()
	*rep = encodeShardMap(&c.shardMap)
	c.mu.Unlock()
}

func MakeMemKVCoordServer(initserver HostName) *MemKVCoord {
	s := new(MemKVCoord)
	s.mu = new(sync.Mutex)

	s.shardMap = make([]HostName, NSHARD)
	for i := uint64(0); i < NSHARD; i++ {
		s.shardMap[i] = initserver
	}
	s.hostShards = make(map[HostName]uint64)
	s.hostShards[initserver] = NSHARD
	s.shardClerks = MakeShardClerkSet()
	return s
}

func (c *MemKVCoord) Start(host HostName) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[COORD_ADD] = func(rawReq []byte, rawRep *[]byte) {
		s := decodeUint64(rawReq)
		c.AddServerRPC(s)
	}
	handlers[COORD_GET] = c.GetShardMapRPC
	s := rpc.MakeRPCServer(handlers)
	s.Serve(host, 1)
}
