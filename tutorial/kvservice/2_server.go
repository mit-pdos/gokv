package kvservice

import (
	"github.com/goose-lang/std"
	"sync"
)

type Server struct {
	mu          *sync.Mutex
	nextFreshId uint64
	lastReplies map[uint64]string

	kvs map[string]string
}

func (s *Server) getFreshNum() uint64 {
	s.mu.Lock()
	n := s.nextFreshId
	s.nextFreshId = std.SumAssumeNoOverflow(s.nextFreshId, 1)
	s.mu.Unlock()
	return n
}

func (s *Server) put(args *putArgs) {
	s.mu.Lock()
	_, ok := s.lastReplies[args.opId]
	if ok {
		s.mu.Unlock()
		return
	}
	s.kvs[args.key] = args.val
	s.lastReplies[args.opId] = ""
	s.mu.Unlock()
}

func (s *Server) conditionalPut(args *conditionalPutArgs) string {
	s.mu.Lock()
	ret, ok := s.lastReplies[args.opId]
	if ok {
		s.mu.Unlock()
		return ret
	}
	if s.kvs[args.key] == args.expectedVal {
		s.kvs[args.key] = args.newVal
		s.lastReplies[args.opId] = ""
	} else {
		s.lastReplies[args.opId] = "ok"
	}
	s.mu.Unlock()
	return ret
}

func (s *Server) get(args *getArgs) string {
	s.mu.Lock()
	ret, ok := s.lastReplies[args.opId]
	if ok {
		s.mu.Unlock()
		return ret
	}
	ret2 := s.kvs[args.key]
	s.mu.Unlock()
	return ret2
}

func MakeServer() *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.kvs = make(map[string]string)
	s.lastReplies = make(map[uint64]string)
	return s
}
