package utproxy

// update-token proxy

import (
	"sync"
)

type UTProxyServer struct {
	mu *sync.Mutex
	// functions to put and get to the actual back-end
	// these should probably be fallible
	putF func(uint64, []byte)
	getF func(uint64, *[]byte)
	// valid update-tokens;
	// updToks : key -> (set of outstanding updToks for that key)
	updToks    map[uint64]map[uint64]bool
	nextUpdTok uint64
}

func (s *UTProxyServer) AcquireUpdateToken(key uint64) uint64 {
	s.mu.Lock()
	u := s.nextUpdTok
	s.nextUpdTok++
	keyUpdToks := s.updToks[key]
	keyUpdToks[u] = true // maps are pointers, so this should be enough
	s.mu.Unlock()
	return u
}

func (s *UTProxyServer) Get(key uint64, cacheable *bool, outv *[]byte) {
	// XXX: we don't want to hold this lock, but I have no idea how to correctly
	// do the back-end get without the lock
	s.mu.Lock()
	if len(s.updToks[key]) == 0 {
		*cacheable = true
	}
	s.mu.Unlock()
	s.getF(key, outv)
}

func (s *UTProxyServer) Put(key uint64, val []byte, updTok uint64) {
	s.mu.Lock()
	if s.updToks[key][updTok] {
		delete(s.updToks[key], updTok)
		if len(s.updToks[key]) == 0 {
			delete(s.updToks, key)
		}
		s.mu.Unlock()
		s.putF(key, val)
	} else {
		s.mu.Unlock()
	}
}
