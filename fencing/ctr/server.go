package ctr

import (
	"sync"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/erpc"
	"github.com/mit-pdos/gokv/fencing/ctr/error_gk"
	"github.com/mit-pdos/gokv/fencing/ctr/getreply_gk"
	"github.com/mit-pdos/gokv/fencing/ctr/putargs_gk"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type Server struct {
	mu *sync.Mutex
	e  *erpc.Server

	v         uint64
	lastEpoch uint64
}

func (s *Server) Put(args *putargs_gk.S) error_gk.E {
	s.mu.Lock()
	// check if epoch is stale
	if args.Epoch < s.lastEpoch {
		s.mu.Unlock()
		return error_gk.EStale
	}

	s.lastEpoch = args.Epoch
	s.v = args.V

	s.mu.Unlock()
	return error_gk.ENone
}

func (s *Server) Get(epoch uint64, reply *getreply_gk.S) {
	s.mu.Lock()
	reply.Err = error_gk.ENone
	// check if epoch is stale
	if epoch < s.lastEpoch {
		s.mu.Unlock()
		reply.Err = error_gk.EStale
		return
	}
	s.lastEpoch = epoch

	reply.Val = s.v
	primitive.Linearize()
	s.mu.Unlock()
	return
}

func StartServer(me grove_ffi.Address) {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.e = erpc.MakeServer()
	s.v = 0

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_GET] = func(raw_args []byte, raw_reply *[]byte) {
		dec := marshal.NewDec(raw_args)
		epoch := dec.GetInt()
		reply := new(getreply_gk.S)
		s.Get(epoch, reply)
		*raw_reply = getreply_gk.Marshal(make([]byte, 0), *reply)
	}

	handlers[RPC_PUT] = s.e.HandleRequest(func(raw_args []byte, reply *[]byte) {
		args, _ := putargs_gk.Unmarshal(raw_args)
		err := s.Put(&args)
		*reply = error_gk.Marshal(make([]byte, 8), err)
	})

	handlers[RPC_FRESHCID] = func(raw_args []byte, reply *[]byte) {
		enc := marshal.NewEnc(8)
		enc.PutInt(s.e.GetFreshCID())
		*reply = enc.Finish()
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
}
