package dir

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
)

type WriteID = uint64

type Value struct {
	// map from chunk index to where the data lives
	servers map[uint64]ChunkHandle
	// TODO: maybe also track length as metadata?
}

type Server struct {
	m *sync.Mutex
	// ongoing can store partial values (some indices are missing)
	ongoing map[WriteID]Value
	// these values are always complete
	data        map[string]Value
	nextWriteId WriteID
}

func StartServer(me grove_ffi.Address) {
	_ = &Server{
		m:           new(sync.Mutex),
		ongoing:     make(map[WriteID]Value),
		data:        make(map[string]Value),
		nextWriteId: 1,
	}
	// TODO: start rpc server
}

// From client
func (s *Server) PrepareWrite() WriteID {
	s.m.Lock()
	id := s.nextWriteId
	s.nextWriteId += 1
	s.ongoing[id] = Value{servers: make(map[uint64]ChunkHandle)}
	s.m.Unlock()
	return id
}

// From chunk
func (s *Server) RecordChunk(writeId WriteID, server grove_ffi.Address, content_hash string,
	index uint64) {
	s.m.Lock()
	// TODO: check if this write is still ongoing
	s.ongoing[writeId].servers[index] = ChunkHandle{Addr: server, ContentHash: content_hash}
	s.m.Unlock()
}

// From chunk
func (s *Server) FinishWrite(writeId WriteID, keyname string) {
	s.m.Lock()
	v := s.ongoing[writeId]
	// TODO: do we want to forget ongoing writes?
	s.data[keyname] = v
	s.m.Unlock()
}
