package lockservice

import (
	"github.com/goose-lang/std"
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

func (s *Server) tryAcquire(id uint64) bool {
	var ret bool
	s.mu.Lock()
	if s.locked {
		ret = (s.holder == id)
	} else {
		if s.holder < id {
			s.holder = id
			s.locked = true
			ret = true
		} else {
			ret = false // may have already been acquired with `id` before
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
	s.mu.Unlock()
}
