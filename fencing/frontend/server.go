package frontend

import (
	// "fmt"
	"github.com/mit-pdos/gokv/fencing/ctr"
	"sync"
)

type GenericKVClerk struct {
	Get func() uint64
	Put func() uint64
}

type Server struct {
	mu   *sync.Mutex
	ck1  *ctr.Clerk
	ctr1 uint64
	ck2  *ctr.Clerk
	ctr2 uint64
}

// pre: key == 0 or key == 1
func (s *Server) FetchAndIncrement(key uint64) {
	s.mu.Lock()
	if key == 0 {
		s.ck1.Put(s.ctr1 + 1)
		s.ctr1 += 1
	} else {
		// key == 1
		s.ck2.Put(s.ctr2 + 1)
		s.ctr2 += 1
	}
	s.mu.Unlock()
}
