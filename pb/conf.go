package pb

import (
	"sync"
)

type VersionedValue struct {
	ver uint64
	val []byte
}

type ConfServer struct {
	mu *sync.Mutex
	kvs map[uint64]VersionedValue
}

func (s *ConfServer)
