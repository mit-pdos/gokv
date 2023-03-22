package dir

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
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

const (
	PrepareWriteId = 1
	RecordChunkId  = 2
	FinishWriteId  = 3
	PrepareReadId  = 4
)

// PrepareWriteArgs is empty

type PreparedWrite struct {
	Id         WriteID
	ChunkAddrs []grove_ffi.Address
}

func ParsePreparedWrite(data []byte) PreparedWrite {
	panic("TODO: marshaling")
}

func MarshalPreparedWrite(id PreparedWrite) []byte {
	panic("TODO: marshaling")
}

type RecordChunkArgs struct {
	WriteId     WriteID
	Server      grove_ffi.Address
	ContentHash string
	Index       uint64
}

func MarshalRecordChunkArgs(args RecordChunkArgs) []byte {
	panic("TODO: marshaling")
}

func ParseRecordChunkArgs(data []byte) RecordChunkArgs {
	panic("TODO: marshaling")
}

type FinishWriteArgs struct {
	WriteId WriteID
	Keyname string
}

func MarshalFinishWriteArgs(args FinishWriteArgs) []byte {
	panic("TODO: marshaling")
}

func ParseFinishWriteArgs(data []byte) FinishWriteArgs {
	panic("TODO: marshaling")
}

type PreparedRead struct {
	Handles []ChunkHandle
}

func MarshalPreparedRead(v PreparedRead) []byte {
	panic("TODO: marshaling")
}

func ParsePreparedRead(data []byte) PreparedRead {
	panic("TODO: marshaling")
}

func StartServer(me grove_ffi.Address) {
	s := &Server{
		m:           new(sync.Mutex),
		ongoing:     make(map[WriteID]Value),
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
		*reply = []byte{}
	}
	handlers[FinishWriteId] = func(req []byte, reply *[]byte) {
		args := ParseFinishWriteArgs(req)
		s.FinishWrite(args)
		*reply = []byte{}
	}
	handlers[PrepareReadId] = func(req []byte, reply *[]byte) {
		args := string(req)
		ret := s.PrepareRead(args)
		*reply = MarshalPreparedRead(ret)
	}
	server := urpc.MakeServer(handlers)
	server.Serve(me)
}

// From client
func (s *Server) PrepareWrite() PreparedWrite {
	s.m.Lock()
	id := s.nextWriteId
	s.nextWriteId += 1
	s.ongoing[id] = Value{servers: make(map[uint64]ChunkHandle)}
	s.m.Unlock()
	return PreparedWrite{
		Id: id,
		// TODO: come up with some chunk servers to return
		// (writes will not work without this)
		ChunkAddrs: []uint64{},
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
	v := s.ongoing[args.WriteId]
	// TODO: do we want to forget ongoing writes?
	s.data[args.Keyname] = v
	s.m.Unlock()
}

func (s *Server) PrepareRead(keyname string) PreparedRead {
	s.m.Lock()
	// need to convert map to slice
	// (the map should be total because it is in s.data)
	indexHandleMap := s.data[keyname].servers
	numChunks := uint64(len(indexHandleMap))
	var servers = make([]ChunkHandle, 0)
	for i := uint64(0); i < numChunks; i++ {
		servers = append(servers, indexHandleMap[i])
	}
	s.m.Unlock()
	return PreparedRead{Handles: servers}
}
