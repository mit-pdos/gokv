package dir

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconnectclient"
)

type Clerk struct {
	client *reconnectclient.ReconnectingClient
}

func MakeClerk(addr grove_ffi.Address) *Clerk {
	client := reconnectclient.MakeReconnectingClient(addr)
	return &Clerk{
		client: client,
	}
}

func (ck *Clerk) PrepareWrite() PreparedWrite {
	reply := new([]byte)
	ck.client.Call(PrepareWriteId, []byte{}, reply, 100 /*ms*/)
	return ParsePreparedWrite(*reply)
}

// From chunk
func (ck *Clerk) RecordChunk(args RecordChunkArgs) {
	req := MarshalRecordChunkArgs(args)
	reply := new([]byte)
	ck.client.Call(RecordChunkId, req, reply, 100 /*ms*/)
}

// From chunk
func (ck *Clerk) FinishWrite(args FinishWriteArgs) {
	req := MarshalFinishWriteArgs(args)
	reply := new([]byte)
	ck.client.Call(FinishWriteId, req, reply, 100 /*ms*/)
}

type ChunkHandle struct {
	Addr        grove_ffi.Address
	ContentHash string
}

func (ck *Clerk) PrepareRead(keyname string) PreparedRead {
	req := []byte(keyname)
	reply := new([]byte)
	ck.client.Call(PrepareReadId, req, reply, 100 /*ms*/)
	return ParsePreparedRead(*reply)
}
