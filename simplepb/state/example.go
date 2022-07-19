package state

import (
	"github.com/mit-pdos/gokv/simplepb/pb"
	"sync"
)

type KVState struct {
	kvs map[uint64][]byte
}

type Op = []byte

func (s *KVState) Apply(op Op) {
	// FIXME: impl
}

type KVServer struct {
	mu *sync.Mutex
	r  *pb.ReplicaServer
}

func (s *KVServer) FetchAndAppend(key uint64, val []byte) {
	s.r.Apply(nil) // FIXME: marshal FAA op
}

func (s *KVServer) SetState(epoch uint64, state []byte) pb.Error {
	s.mu.Lock()
	// if s.epoch >= epoch {
	// return pb.EStale
	// }
	// FIXME: set state by unmarshalling
	s.mu.Unlock()
	return pb.ENone
}

func (s *KVServer) GetState(epoch uint64) (pb.Error, []byte) {
	if s.r.EpochFence(epoch) {
		return pb.EStale, nil
	}

	s.mu.Lock()
	// FIXME: marshal state
	s.mu.Unlock()
	return pb.ENone, nil
}
