package ctr

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu sync.Mutex
	v  uint64

	lastSeq   map[uint64]uint64
	lastReply map[uint64]uint64
	lastCID   uint64
}

func (s *Server) Put(args *PutArgs) {
	s.mu.Lock()

	last, ok := s.lastSeq[args.cid]
	seq := args.seq
	if ok && seq <= last {
		return
	}

	s.v = args.v
	s.lastSeq[args.cid] = seq

	s.mu.Unlock()
}

func (s *Server) Get() uint64 {
	s.mu.Lock()
	ret := s.v
	s.mu.Unlock()
	return ret
}

func (s *Server) GetFreshCID() uint64 {
	s.mu.Lock()
	s.lastCID += 1
	ret := s.lastCID
	s.mu.Unlock()
	return ret
}

func StartServer(me grove_ffi.Address) {
	s := new(Server)
	s.lastCID = 0
	s.v = 0
	s.lastSeq = make(map[uint64]uint64)
	s.lastReply = make(map[uint64]uint64)

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_GET] = func(args []byte, reply *[]byte) {
		enc := marshal.NewEnc(8)
		enc.PutInt(s.Get())
		*reply = enc.Finish()
	}

	handlers[RPC_PUT] = func(raw_args []byte, reply *[]byte) {
		args := DecPutArgs(raw_args)
		s.Put(args)
		// FIXME: this might return an error saying that the Put was from an old
		// epoch
	}

	handlers[RPC_FRESHCID] = func(raw_args []byte, reply *[]byte) {
		enc := marshal.NewEnc(8)
		enc.PutInt(s.GetFreshCID())
		*reply = enc.Finish()
	}

	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)
}
