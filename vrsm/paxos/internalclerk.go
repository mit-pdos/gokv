package paxos

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconnectclient"
	"github.com/mit-pdos/gokv/vrsm/paxos/applyasfollowerargs_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/applyasfollowerreply_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/enternewepochargs_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/enternewepochreply_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/error_gk"
)

const (
	RPC_APPLY_AS_FOLLOWER = uint64(0)
	RPC_ENTER_NEW_EPOCH   = uint64(1)
	RPC_BECOME_LEADER     = uint64(2)
)

// these clerks hide connection failures, and retry forever
type singleClerk struct {
	cl *reconnectclient.ReconnectingClient
}

func MakeSingleClerk(addr grove_ffi.Address) *singleClerk {
	// make a bunch of urpc clients
	ck := &singleClerk{
		cl: reconnectclient.MakeReconnectingClient(addr),
	}

	return ck
}

func (s *singleClerk) enterNewEpoch(args *enternewepochargs_gk.S) *enternewepochreply_gk.S {
	raw_args := enternewepochargs_gk.Marshal(make([]byte, 0), *args)
	raw_reply := new([]byte)
	err := s.cl.Call(RPC_ENTER_NEW_EPOCH, raw_args, raw_reply, 500 /* ms */)
	if err == 0 {
		ret, _ := enternewepochreply_gk.Unmarshal(*raw_reply)
		return &ret
	} else {
		return &enternewepochreply_gk.S{Err: error_gk.ETimeout}
	}
}

func (s *singleClerk) applyAsFollower(args *applyasfollowerargs_gk.S) *applyasfollowerreply_gk.S {
	raw_args := applyasfollowerargs_gk.Marshal(make([]byte, 0), *args)
	raw_reply := new([]byte)
	err := s.cl.Call(RPC_APPLY_AS_FOLLOWER, raw_args, raw_reply, 500 /* ms */)
	if err == 0 {
		ret, _ := applyasfollowerreply_gk.Unmarshal(*raw_reply)
		return &ret
	} else {
		return &applyasfollowerreply_gk.S{Err: error_gk.ETimeout}
	}
}

func (s *singleClerk) TryBecomeLeader() {
	// make the server the primary
	reply := new([]byte)
	s.cl.Call(RPC_BECOME_LEADER, make([]byte, 0), reply, 500 /* ms */)
}
