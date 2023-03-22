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
)

func ParseWriteID(data []byte) WriteID {
	panic("TODO: marshaling")
}

func MarshalWriteID(id WriteID) []byte {
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
		*reply = MarshalWriteID(ret)
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
	server := urpc.MakeServer(handlers)
	server.Serve(me)
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
