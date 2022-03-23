package ctr

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu        *sync.Mutex
	v         uint64
	lastEpoch uint64

	lastSeq   map[uint64]uint64
	lastReply map[uint64]uint64
	lastCID   uint64
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

	// check if we've seen this request before
	last, ok := s.lastSeq[args.cid]
	seq := args.seq
	if ok && seq <= last {
		s.mu.Unlock()
		return ENone
	}

	s.v = args.v
	s.lastSeq[args.cid] = seq

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

func (s *Server) GetFreshCID() uint64 {
	s.mu.Lock()
	s.lastCID += 1
	ret := s.lastCID
	s.mu.Unlock()
	return ret
}

func StartServer(me grove_ffi.Address) {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.lastCID = 0
	s.v = 0
	s.lastSeq = make(map[uint64]uint64)
	s.lastReply = make(map[uint64]uint64)

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_GET] = func(raw_args []byte, raw_reply *[]byte) {
		args := DecGetArgs(raw_args)
		reply := new(GetReply)
		s.Get(args, reply)
		*raw_reply = EncGetReply(reply)
	}

	handlers[RPC_PUT] = func(raw_args []byte, reply *[]byte) {
		args := DecPutArgs(raw_args)
		err := s.Put(args)
		enc := marshal.NewEnc(8)
		enc.PutInt(err)
		*reply = enc.Finish()
	}

	handlers[RPC_FRESHCID] = func(raw_args []byte, reply *[]byte) {
		enc := marshal.NewEnc(8)
		enc.PutInt(s.GetFreshCID())
		*reply = enc.Finish()
	}

	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)
}
