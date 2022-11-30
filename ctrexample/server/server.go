package main

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
	"sync"
)

type CtrServer struct {
	mu  *sync.Mutex
	val uint64

	filename string
}

// requires lock to be held
func (s *CtrServer) MakeDurable() {
	e := marshal.NewEnc(8)
	e.PutInt(s.val)
	grove_ffi.FileWrite(s.filename, e.Finish())
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

	a := grove_ffi.FileRead(s.filename)
	if len(a) == 0 {
		s.val = 0
	} else {
		d := marshal.NewDec(a)
		s.val = d.GetInt()
	}

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[0] = func(args []byte, reply *[]byte) {
		v := s.FetchAndIncrement()
		e := marshal.NewEnc(8)
		e.PutInt(v)
		*reply = e.Finish()
	}
	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
	// select{}
}
