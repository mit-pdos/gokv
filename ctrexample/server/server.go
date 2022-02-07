package main

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/goose/machine"
	"sync"
)

type CtrServer struct {
	mu       *sync.Mutex
	val      uint64

	filename string
}

// requires lock to be held
func (s *CtrServer) MakeDurable() {
	a := make([]byte, 8)
	machine.UInt64Put(a, s.val)
	grove_ffi.Write(s.filename, a)
}

func (s *CtrServer) FetchAndIncrement() uint64 {
	s.mu.Lock()
	ret := s.val
	s.val += 1
	s.MakeDurable()
	s.mu.Unlock()
	return ret
}

// the boot/main() function for the server
func main() {
	me := uint64(53021371269120) // hard-coded "127.0.0.1:12345"
	s := new(CtrServer)
	s.mu = new(sync.Mutex)
	s.filename = "ctr"

	a := grove_ffi.Read(s.filename)
	if len(a) == 0 {
		s.val = 0
	} else {
		s.val = machine.UInt64Get(a)
	}

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[0] = func(args []byte, reply *[]byte) {
		v := s.FetchAndIncrement()
		*reply = make([]byte, 8)
		machine.UInt64Put(*reply, v)
	}
	rs := rpc.MakeRPCServer(handlers)
	rs.Serve(me, 1)
	// select{}
}
