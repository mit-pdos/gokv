package chunk

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/trusted_hash"
	"github.com/mit-pdos/gokv/tutorial/objectstore/dir"
)

type Server struct {
	m      *sync.Mutex
	chunks map[string][]byte
	dir    *dir.Clerk
	me     grove_ffi.Address
}

func StartServer(me grove_ffi.Address, dir_addr grove_ffi.Address) {
	dir := dir.MakeClerk(dir_addr)
	_ = &Server{
		m:      new(sync.Mutex),
		chunks: make(map[string][]byte),
		dir:    dir,
		me:     me,
	}
	// TODO: start rpc server
	/*
		handlers := make(map[uint64]func([]byte, *[]byte))
		handlers[GetDecisionId] = func(_req []byte, reply *[]byte) {
			decision := coordinator.GetDecision()
			replyData := make([]byte, 1)
			replyData[0] = decision
			*reply = replyData
		}
		server := urpc.MakeServer(handlers)
		server.Serve(me)
	*/

}

func (s *Server) WriteChunk(writeId WriteID, chunk []byte, index uint64) {
	content_hash := trusted_hash.Hash(chunk)
	s.m.Lock()
	s.chunks[content_hash] = chunk
	s.m.Unlock()
	s.dir.RecordChunk(writeId, s.me, content_hash, index)
}

func (s *Server) GetChunk(content_hash string) []byte {
	s.m.Lock()
	data := s.chunks[content_hash]
	s.m.Unlock()
	return data
}
