package ftconfig

import (
	"github.com/mit-pdos/gokv/simplepb/e"
	"sync"
)

type Epoch struct {
	epoch uint64
	id    uint64
}

func EpochLt(lhs Epoch, rhs Epoch) bool {
	return (lhs.epoch < rhs.epoch) || (lhs.epoch == rhs.epoch && lhs.id < rhs.id)
}

func EpochLe(lhs Epoch, rhs Epoch) bool {
	return (lhs == rhs) || EpochLt(lhs, rhs)
}

type FollowerServer struct {
	mu            *sync.Mutex
	epoch         Epoch
	acceptedEpoch Epoch
	config        []uint64

	// these two together form the "value"
	nextConfig []uint64
	state      []byte

	nextIndex uint64

	isLeader  bool
	committed bool
}

func (s *FollowerServer) commitOld() e.Error {
	s.mu.Lock()
	if s.committed {
		s.mu.Unlock()
		return e.None
	}

	s.mu.Unlock()
	return e.None
}

func (s *FollowerServer) BecomeFollower(args *BecomeFollowerArgs, reply *BecomeFollowerReply) {
	s.mu.Lock()
	if EpochLe(args.epoch, s.epoch) {
		s.mu.Unlock()
		reply.err = e.Stale
		return
	}
	// enter new epoch
	s.epoch = args.epoch

	reply.err = e.None
	reply.nextIndex = s.nextIndex
	reply.state = s.state
	reply.nextConfig = s.nextConfig
	reply.config = s.config
	reply.epoch = s.acceptedEpoch

	s.mu.Unlock()
}

func (s *FollowerServer) ConfigChange(newConfig []uint64) e.Error {
	s.mu.Lock()
	s.nextIndex += 1

	s.mu.Unlock()
	return e.None
}

func (s *FollowerServer) ApplyAsFollower(args *ApplyAsFollowerArgs) e.Error {
	s.mu.Lock()
	if EpochLt(args.epoch, s.epoch) {
		s.mu.Unlock()
		return e.Stale
	}

	if EpochLt(s.epoch, args.epoch) || args.nextIndex > s.nextIndex {
		s.epoch = args.epoch
		s.state = args.state
		s.nextIndex = args.nextIndex
		s.acceptedEpoch = args.epoch
	}
	s.mu.Unlock()
	return e.None
}
