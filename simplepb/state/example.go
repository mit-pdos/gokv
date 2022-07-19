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

func (s *KVState) GetState() {
	// FIXME: impl
}

func (s *KVState) SetState(snap []byte) {
	// FIXME: impl
}

type KVServer struct {
	r  *pb.ReplicaServer
}

func (s *KVServer) FetchAndAppend(key uint64, val []byte) {
	s.r.Apply(nil) // FIXME: marshal FAA op
}

func (s *KVServer) FetchAndAppend(key uint64, val []byte) {
	s.r.Apply(nil) // FIXME: marshal FAA op
}
