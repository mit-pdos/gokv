package kvcached

import (
	"sync"
)

type KVCacheServer struct {
	mu      *sync.Mutex
	kvs     map[uint64][]byte
	kleases map[uint64]uint64
	nextlid uint64
}

// Does a put only if the lease is valid
func (s *KVCacheServer) PutRPC(key uint64, lid uint64, val []byte) {
	s.mu.Lock()
	// check that lease is still valid
	if s.kleases[key] >= lid {
		s.mu.Unlock()
		return
	}
	delete(s.kleases, key)
	s.kvs[key] = val
	s.mu.Unlock()
}

// returns true iff the key existed in the map
func (s *KVCacheServer) GetRPC(key uint64, existed *bool, lid *uint64, outv *[]byte) {
	s.mu.Lock()
	v, ok := s.kvs[key]
	if !ok {
		*lid = s.nextlid
		s.kleases[key] = s.nextlid
	}
	s.nextlid++
	s.mu.Unlock()
	*outv = v
	*existed = ok
}
