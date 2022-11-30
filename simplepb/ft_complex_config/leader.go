package ftconfig

import (
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/tchajed/goose/machine"
	"sync"
)

type StateMachine struct {
	Apply func(val []byte, op []byte) ([]byte, []byte)
}

type LeaderServer struct {
	mu     *sync.Mutex
	id     uint64
	epoch  Epoch
	config []uint64
	clerks []*FollowerClerk

	// these two together form the "value"
	nextConfig []uint64
	state      []byte

	nextIndex uint64
	sm        *StateMachine

	isLeader  bool
	committed bool
}

func confEq(c1 []uint64, c2 []uint64) bool {
	if len(c1) != len(c2) {
		return false
	} else {
		var eq = true
		for i, c := range c1 {
			if c != c2[i] {
				eq = false
				break
			}
		}
		return eq
	}
}

func max(a uint64, b uint64) uint64 {
	if a < b {
		return b
	} else {
		return a
	}
}

func (s *LeaderServer) TryBecomeLeader() e.Error {
	s.mu.Lock()
	// Make a single GetEpoch RPC to get the epoch number of a random old server...
	randi := machine.RandomUint64()
	err, oldepoch := s.clerks[randi].GetEpoch()
	if err != e.None {
		s.mu.Unlock()
		return err
	}
	// ... and try to pick something bigger.
	newepoch := Epoch{id: s.id, epoch: max(s.epoch.epoch, oldepoch.epoch+1)}

	// Make BecomeFollower RPCs
	// collect votes and responses from what we guess is the latest config
	args := &BecomeFollowerArgs{epoch: newepoch}
	clerks := s.clerks
	config := s.config
	mu := new(sync.Mutex)
	replies := make([]*BecomeFollowerReply, len(clerks))
	s.mu.Unlock()

	for i, ck := range clerks {
		ck := ck
		i := i
		go func() {
			// XXX: using tmpreply to avoid race with the later code that reads
			// the replies
			tmpreply := new(BecomeFollowerReply)
			ck.BecomeFollower(args, tmpreply)

			mu.Lock()
			replies[i] = tmpreply
			mu.Unlock()
		}()
	}

	machine.Sleep(100 * 1_000_000)
	// XXX: this waits only 100 ms; if the median RPC latency to the old
	// config's followers is more than this, then we will timeout too early.

	var latestReply *BecomeFollowerReply
	latestReply = replies[0]
	var numSuccessful uint64 = 0

	// analyze the replies
	mu.Lock()
	for _, reply := range replies {
		if reply.err == e.None {
			numSuccessful += 1

			if EpochLt(latestReply.epoch, reply.epoch) {
				latestReply = reply
			} else if latestReply.epoch == reply.epoch {
				if latestReply.nextIndex < reply.nextIndex {
					latestReply = reply
				}
			}
		}
	}
	s.mu.Lock()
	if EpochLt(s.epoch, args.epoch) {
		// if the
		if confEq(config, latestReply.config) {
			s.epoch = args.epoch
			s.nextIndex = latestReply.nextIndex
			s.state = latestReply.state
			// s.config = latestReply.config
			s.nextConfig = latestReply.nextConfig
		}
	}
	s.mu.Unlock()
	mu.Unlock() // XXX: need to release this because some of the goroutines
	// making RPCs might be waiting for this still

	return e.None
}

func (s *LeaderServer) commitOld() e.Error {
	s.mu.Lock()
	if s.committed {
		s.mu.Unlock()
		return e.None
	}

	s.mu.Unlock()
	return e.None
}

func (s *LeaderServer) Apply(op []byte) (e.Error, []byte) {
	var reply []byte
	s.mu.Lock()
	if !s.isLeader {
		s.mu.Unlock()
		return e.NotLeader, nil
	}

	for !s.committed {
		s.mu.Unlock()
		err := s.commitOld()
		if err != e.None {
			return err, nil
		}
		s.mu.Lock()
	}

	s.state, reply = s.sm.Apply(s.state, op)
	oldConfig := s.config
	s.config = s.nextConfig
	s.nextIndex += 1
	s.committed = false

	// if the next entry (that we want to propose+commit) is in a different
	// config, we have to rerun phase1. So, mark ourselves as no longer a leader.
	if !confEq(oldConfig, s.nextConfig) {
		s.isLeader = false
		s.mu.Unlock()
		return e.NotLeader, nil
	}

	s.mu.Unlock()
	// finished creating proposal

	// FIXME: send proposal to all servers in config, and wait for a majority of them
	// to reply.

	return e.None, reply
}
