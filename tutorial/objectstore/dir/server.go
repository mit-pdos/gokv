package dir

import "github.com/mit-pdos/gokv/grove_ffi"

type WriteID = uint64

type Server struct {
}

// From client
func (s *Server) PrepareWrite() WriteID {
	panic("impl")
}

// From chunk
func (s *Server) RecordChunk(writeId WriteID, server grove_ffi.Address, content_hash uint64,
	index uint64) {
	panic("impl")
}

// From chunk
func (s *Server) FinishWrite(writeId WriteID, keyname string) {
}
