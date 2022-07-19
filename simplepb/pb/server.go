package pb

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
)

type Op = []byte

type StateMachine struct {
	Apply    func(op Op) []byte
	SetState func(snap []byte)
	GetState func() []byte
}

type ReplicaServer struct {
	mu        *sync.Mutex
	epoch     uint64
	sm        *StateMachine
	nextIndex uint64 // this is a per-epoch deduplication ID
	// reset to 0 when entering a new epoch

	isPrimary bool
	clerks    []*Clerk
}

type Error = uint64

const (
	ENone       = uint64(0)
	EStale      = uint64(1)
	EOutOfOrder = uint64(2)
)

// called on the primary server to apply a new operation.
func (s *ReplicaServer) Apply(op Op) (Error, []byte) {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return EStale, nil
	}

	// apply it locally
	ret := s.sm.Apply(op)

	nextIndex := s.nextIndex
	epoch := s.epoch
	clerks := s.clerks
	s.mu.Unlock()

	// tell backups to apply it
	wg := new(sync.WaitGroup)
	errs := make([]Error, len(clerks))
	for i, clerk := range clerks {
		clerk := clerk
		i := i
		wg.Add(1)
		go func() {
			errs[i] = clerk.Apply(epoch, nextIndex, op)
		}()
	}
	wg.Wait()
	var err = ENone
	for _, e := range errs {
		if e != ENone {
			err = e
		}
	}

	return err, ret
}

// called on backup servers to apply an operation so it is replicated and
// can be considered committed by primary.
func (s *ReplicaServer) ApplyAsBackup(epoch uint64, index uint64, op Op) Error {
	s.mu.Lock()
	if s.EpochFence(epoch) {
		s.mu.Unlock()
		return EStale
	}

	if index != s.nextIndex {
		s.mu.Unlock()
		return EOutOfOrder
	}

	// apply it locally
	s.sm.Apply(op)
	s.nextIndex += 1

	s.mu.Unlock()
	return ENone
}

func (s *ReplicaServer) SetState(epoch uint64, state []byte) Error {
	s.mu.Lock()
	if s.epoch >= epoch {
		return EStale
	}

	s.sm.SetState(state)

	s.mu.Unlock()
	return ENone
}

func (s *ReplicaServer) GetState(epoch uint64) (Error, []byte) {
	s.mu.Lock()
	if s.EpochFence(epoch) {
		s.mu.Unlock()
		return EStale, nil
	}

	ret := s.sm.GetState()
	s.mu.Unlock()

	return ENone, ret
}

// returns true iff stale
func (s *ReplicaServer) EpochFence(epoch uint64) bool {
	if s.epoch < epoch {
		s.epoch = epoch
		s.isPrimary = false
		s.nextIndex = 0
	}
	return s.epoch > epoch
}

func (s *ReplicaServer) BecomePrimary(epoch uint64, replicas []grove_ffi.Address) Error {
	s.mu.Lock()
	if s.EpochFence(epoch) {
		s.mu.Unlock()
		return EStale
	}
	s.isPrimary = true

	s.clerks = make([]*Clerk, len(replicas))
	for i := range s.clerks {
		s.clerks[i] = MakeClerk(replicas[i])
	}

	s.mu.Unlock()
	return ENone
}
