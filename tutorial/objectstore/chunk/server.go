package chunk

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/trusted_hash"
	"github.com/mit-pdos/gokv/tutorial/objectstore/chunk/writechunk_gk"
	"github.com/mit-pdos/gokv/tutorial/objectstore/dir"
	"github.com/mit-pdos/gokv/tutorial/objectstore/dir/recordchunk_gk"
	"github.com/mit-pdos/gokv/urpc"
)

type Server struct {
	m      *sync.Mutex
	chunks map[string][]byte
	dir    *dir.Clerk
	me     grove_ffi.Address
}

func (s *Server) WriteChunk(args writechunk_gk.S) {
	content_hash := trusted_hash.Hash(args.Chunk)
	s.m.Lock()
	s.chunks[content_hash] = args.Chunk
	s.m.Unlock()
	s.dir.RecordChunk(recordchunk_gk.S{
		WriteId:     args.WriteId,
		Server:      s.me,
		ContentHash: content_hash,
		Index:       args.Index,
	})
}

func (s *Server) GetChunk(content_hash string) []byte {
	s.m.Lock()
	data := s.chunks[content_hash]
	s.m.Unlock()
	return data
}

func StartServer(me grove_ffi.Address, dir_addr grove_ffi.Address) {
	dir := dir.MakeClerk(dir_addr)
	s := &Server{
		m:      new(sync.Mutex),
		chunks: make(map[string][]byte),
		dir:    dir,
		me:     me,
	}
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[WriteChunkId] = func(req []byte, reply *[]byte) {
		args, _ := writechunk_gk.Unmarshal(req)
		s.WriteChunk(args)
		*reply = make([]byte, 0) // TODO: is this needed?
	}
	handlers[GetChunkId] = func(req []byte, reply *[]byte) {
		// inline marshaling because types are so simple
		args := string(req)
		ret := s.GetChunk(args)
		*reply = ret
	}
	server := urpc.MakeServer(handlers)
	server.Serve(me)
}
