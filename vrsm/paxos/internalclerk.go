package paxos

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconnectclient"
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

func (s *singleClerk) enterNewEpoch(args *enterNewEpochArgs) *enterNewEpochReply {
	raw_args := encodeEnterNewEpochArgs(args)
	raw_reply := new([]byte)
	err := s.cl.Call(RPC_ENTER_NEW_EPOCH, raw_args, raw_reply, 500 /* ms */)
	if err == 0 {
		return decodeEnterNewEpochReply(*raw_reply)
	} else {
		return &enterNewEpochReply{err: ETimeout}
	}

}

func (s *singleClerk) applyAsFollower(args *applyAsFollowerArgs) *applyAsFollowerReply {
	raw_args := encodeApplyAsFollowerArgs(args)
	raw_reply := new([]byte)
	err := s.cl.Call(RPC_APPLY_AS_FOLLOWER, raw_args, raw_reply, 500 /* ms */)
	if err == 0 {
		return decodeApplyAsFollowerReply(*raw_reply)
	} else {
		return &applyAsFollowerReply{err: ETimeout}
	}
}

func (s *singleClerk) TryBecomeLeader() {
	// make the server the primary
	reply := new([]byte)
	s.cl.Call(RPC_BECOME_LEADER, make([]byte, 0), reply, 500 /* ms */)
}
