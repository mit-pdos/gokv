package lockservice

import (
	"github.com/goose-lang/std"
	"log"
	"sync"
)

type Server struct {
	mu     *sync.Mutex
	nextId uint64
	locked bool
	holder uint64
}

func (s *Server) getFreshNum() uint64 {
	s.mu.Lock()
	n := s.nextId
	s.nextId = std.SumAssumeNoOverflow(s.nextId, 1)
	s.mu.Unlock()
	return n
}

const (
	StatusGranted = uint64(0)
	StatusRetry   = uint64(2)
	StatusStale   = uint64(1)
)

func (s *Server) tryAcquire(id uint64) uint64 {
	var ret uint64
	s.mu.Lock()
	if s.holder > id {
		ret = StatusStale
	} else {
		if s.locked {
			if s.holder == id {
				ret = StatusGranted
			} else {
				ret = StatusRetry
			}
		} else {
			s.holder = id
			s.locked = true
			log.Printf("Lock held by %d", id)
			ret = StatusGranted
		}
	}
	s.mu.Unlock()
	return ret
}

func (s *Server) release(id uint64) {
	s.mu.Lock()
	if s.holder == id {
		s.locked = false
	}
	log.Printf("Lock released by %d", id)
	s.mu.Unlock()
}

func MakeServer() *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	return s
}
