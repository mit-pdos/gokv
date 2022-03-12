package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu           *sync.Mutex
	data         grove_ffi.Address
	currentEpoch uint64
}

func (s *Server) Set(newFrontend grove_ffi.Address) uint64 {
	s.mu.Lock()
	s.data = newFrontend
	s.currentEpoch += 1
	ret := s.currentEpoch
	s.mu.Unlock()
	return ret
}

// XXX: don't need to send fencing token here, because client won't need it
func (s *Server) Get() grove_ffi.Address {
	s.mu.Lock()
	ret := s.data
	s.mu.Unlock()
	return ret
}

func StartServer(me grove_ffi.Address) {
	s := new(Server)
	s.data = 0
	s.currentEpoch = 0

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_SET] = func(args []byte, reply *[]byte) {
		dec := marshal.NewDec(args)
		enc := marshal.NewEnc(8)
		enc.PutInt(s.Set(dec.GetInt()))
		*reply = enc.Finish()
	}

	handlers[RPC_GET] = func(args []byte, reply *[]byte) {
		enc := marshal.NewEnc(8)
		enc.PutInt(s.Get())
		*reply = enc.Finish()
	}

	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)
}
