package index

import (
	"github.com/mit-pdos/go-mvcc/index"
)

type Server struct {
	index *index.Index
}

func (s *Server) AcquireTuple(key uint64, tid uint64) uint64 {
	return s.index.GetTuple(key).Own(tid)
}

func (s *Server) Read(key uint64, tid uint64) string {
	t := s.index.GetTuple(key)
	t.ReadWait(tid)
	val, _ := s.index.GetTuple(key).ReadVersion(tid)
	return val
}

func (s *Server) UpdateAndRelease(tid uint64, writes map[uint64]string) {
	for key, val := range writes {
		t := s.index.GetTuple(key)
		t.WriteLock()
		t.AppendVersion(tid, val)
	}
}

func MakeServer() *Server {
	return &Server{
		index: index.MkIndex(),
	}
}
