package ctr

import (
	"github.com/mit-pdos/gokv/erpc"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu *sync.Mutex
	e  *erpc.Server

	v         uint64
	lastEpoch uint64
}

const (
	ENone  = uint64(0)
	EStale = uint64(1)
)

func (s *Server) Put(args *PutArgs) uint64 {
	s.mu.Lock()
	// check if epoch is stale
	if args.epoch < s.lastEpoch {
		s.mu.Unlock()
		return EStale
	}

	s.lastEpoch = args.epoch
	s.v = args.v

	s.mu.Unlock()
	return ENone
}

func (s *Server) Get(args *GetArgs, reply *GetReply) {
	s.mu.Lock()
	reply.err = ENone
	// check if epoch is stale
	if args.epoch < s.lastEpoch {
		s.mu.Unlock()
		reply.err = EStale
		return
	}
	s.lastEpoch = args.epoch

	// XXX: for the proof, we're going to have to use the reply table here.
	// Hopefully, prophecy variables can one day fix that.

	reply.val = s.v
	s.mu.Unlock()
	return
}

func StartServer(me grove_ffi.Address) {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.e = erpc.MakeServer()
	s.v = 0

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_GET] = s.e.HandleRequest(func(raw_args []byte, raw_reply *[]byte) {
		args := DecGetArgs(raw_args)
		reply := new(GetReply)
		s.Get(args, reply)
		*raw_reply = EncGetReply(reply)
	})

	handlers[RPC_PUT] = s.e.HandleRequest(func(raw_args []byte, reply *[]byte) {
		args := DecPutArgs(raw_args)
		err := s.Put(args)
		enc := marshal.NewEnc(8)
		enc.PutInt(err)
		*reply = enc.Finish()
	})

	handlers[RPC_FRESHCID] = func(raw_args []byte, reply *[]byte) {
		enc := marshal.NewEnc(8)
		enc.PutInt(s.e.GetFreshCID())
		*reply = enc.Finish()
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
}
