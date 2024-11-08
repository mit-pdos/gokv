package kvservice

import (
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/tutorial/kvservice/conditionalput_gk"
	"github.com/mit-pdos/gokv/tutorial/kvservice/get_gk"
	"github.com/mit-pdos/gokv/tutorial/kvservice/put_gk"
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

func (s *Server) put(args *put_gk.S) {
	s.mu.Lock()
	_, ok := s.lastReplies[args.OpId]
	if ok {
		s.mu.Unlock()
		return
	}
	s.kvs[args.Key] = args.Val
	s.lastReplies[args.OpId] = ""
	s.mu.Unlock()
}

func (s *Server) conditionalPut(args *conditionalput_gk.S) string {
	s.mu.Lock()
	ret, ok := s.lastReplies[args.OpId]
	if ok {
		s.mu.Unlock()
		return ret
	}

	var ret2 string = ""
	if s.kvs[args.Key] == args.ExpectedVal {
		s.kvs[args.Key] = args.NewVal
		ret2 = "ok"
	}
	s.lastReplies[args.OpId] = ret2
	s.mu.Unlock()
	return ret2
}

func (s *Server) get(args *get_gk.S) string {
	s.mu.Lock()
	ret, ok := s.lastReplies[args.OpId]
	if ok {
		s.mu.Unlock()
		return ret
	}
	ret2 := s.kvs[args.Key]
	s.lastReplies[args.OpId] = ret2
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
