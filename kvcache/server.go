package kvcache

import (
	"sync"
)

type VersionedValue struct {
	vnum uint64
	val  []byte
}

type KVCacheServer struct {
	mu          *sync.Mutex
	kvs         map[uint64]VersionedValue
	leases      map[uint64]uint64
	nextVersion uint64
}

// Does a put only if the version number is larger than the highest vnum seen
// before
func (s *KVCacheServer) PutRPC(key uint64, vnum uint64, val []byte) {
	s.mu.Lock()
	s.mu.Unlock()
}

// returns true iff the key existed in the map
func (s *KVCacheServer) GetRPC(key uint64, outv *[]byte) bool {
	s.mu.Lock()
	v, ok := s.kvs[key]
	*outv = v.val
	s.mu.Unlock()
	if ok {
		return true
	} else {
		s.mu.Unlock()
		return false
	}
}

// returns true iff the key existed in the map;
// if it didn't exist, this returns a lease on the key
func (s *KVCacheServer) GetWithLeaseRPC(key uint64, outv *[]byte) bool {
	s.mu.Lock()
	v, ok := s.kvs[key]
	*outv = v.val
	s.mu.Unlock()
	if ok {
		return true
	} else {
		s.mu.Unlock()
		return false
	}
}
