package dir

import "github.com/mit-pdos/gokv/grove_ffi"

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

type ChunkHandle struct {
	Addr        grove_ffi.Address
	ContentHash string
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
