package dir

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type PartialValue struct {
	// map from chunk index to where the data lives
	servers map[uint64]ChunkHandle
	// TODO: maybe also track length as metadata?
}

type Value struct {
	// list of where each chunk lives
	servers []ChunkHandle
}

type Server struct {
	m *sync.Mutex
	// ongoing can store partial values (some indices are missing)
	ongoing     map[WriteID]PartialValue
	data        map[string]Value
	nextWriteId WriteID
}

// From client
func (s *Server) PrepareWrite() PreparedWrite {
	s.m.Lock()
	id := s.nextWriteId
	s.nextWriteId += 1
	s.ongoing[id] = PartialValue{servers: make(map[uint64]ChunkHandle)}
	s.m.Unlock()
	return PreparedWrite{
		Id: id,
		// TODO: come up with some chunk servers to return
		// (writes will not work without this)
		ChunkAddrs: make([]uint64, 0),
	}
}

// From chunk
func (s *Server) RecordChunk(args RecordChunkArgs) {
	s.m.Lock()
	// TODO: check if this write is still ongoing
	s.ongoing[args.WriteId].servers[args.Index] = ChunkHandle{
		Addr:        args.Server,
		ContentHash: args.ContentHash,
	}
	s.m.Unlock()
}

// From chunk
func (s *Server) FinishWrite(args FinishWriteArgs) {
	s.m.Lock()
	v := s.ongoing[args.WriteId].servers
	// TODO: do we want to forget ongoing writes?

	numChunks := uint64(len(v))
	var servers = make([]ChunkHandle, 0)
	for i := uint64(0); i < numChunks; i++ {
		servers = append(servers, v[i])
	}

	s.data[args.Keyname] = Value{servers: servers}
	s.m.Unlock()
}

func (s *Server) PrepareRead(keyname string) PreparedRead {
	s.m.Lock()
	// need to convert map to slice
	// (the map should be total because it is in s.data)
	servers := s.data[keyname].servers
	s.m.Unlock()
	return PreparedRead{Handles: servers}
}

func StartServer(me grove_ffi.Address) {
	s := &Server{
		m:           new(sync.Mutex),
		ongoing:     make(map[WriteID]PartialValue),
		data:        make(map[string]Value),
		nextWriteId: 1,
	}
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[PrepareWriteId] = func(_req []byte, reply *[]byte) {
		ret := s.PrepareWrite()
		*reply = MarshalPreparedWrite(ret)
	}
	handlers[RecordChunkId] = func(req []byte, reply *[]byte) {
		args := ParseRecordChunkArgs(req)
		s.RecordChunk(args)
		*reply = make([]byte, 0)
	}
	handlers[FinishWriteId] = func(req []byte, reply *[]byte) {
		args := ParseFinishWriteArgs(req)
		s.FinishWrite(args)
		*reply = make([]byte, 0)
	}
	handlers[PrepareReadId] = func(req []byte, reply *[]byte) {
		args := string(req)
		ret := s.PrepareRead(args)
		*reply = MarshalPreparedRead(ret)
	}
	server := urpc.MakeServer(handlers)
	server.Serve(me)
}
