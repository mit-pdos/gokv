package minlease

import (
	"sync"

	"github.com/goose-lang/primitive"
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
)

const (
	RPC_GET = uint64(0)
	RPC_PUT = uint64(1)
)

type Server struct {
	mu              *sync.Mutex
	val             uint64
	leaseExpiration uint64
}

// return true iff successful
func (s *Server) TryLocalIncrement() bool {
	s.mu.Lock()
	_, h := grove_ffi.GetTimeRange()
	if h >= s.leaseExpiration {
		s.mu.Unlock()
		return false
	}
	s.val = std.SumAssumeNoOverflow(s.val, 1)
	s.mu.Unlock()
	return true
}

func (s *Server) Put(val uint64) {
	s.mu.Lock()
	s.val = val
	s.mu.Unlock()
}

func (s *Server) Get() uint64 {
	s.mu.Lock()
	v := s.val
	s.mu.Unlock()
	return v
}

func StartServer() *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.val = 0
	s.leaseExpiration = 10 // FIXME: put a reasonable time for the lease expiration

	go func() {
		for s.TryLocalIncrement() {
		}
	}()

	return s
}

func client(s *Server) {
	for {
		l, _ := grove_ffi.GetTimeRange()
		if l > s.leaseExpiration {
			break
		}
		primitive.Sleep(s.leaseExpiration - l)
	}

	// now the client has the lease
	v := s.Get()
	newv := std.SumAssumeNoOverflow(v, 1)
	s.Put(newv)
	v2 := s.Get()
	primitive.Assert(v2 == newv)
}

func main() {
	s := StartServer()
	client(s)
}
