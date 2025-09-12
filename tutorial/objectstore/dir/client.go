package dir

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconnectclient"
	"github.com/mit-pdos/gokv/tutorial/objectstore/dir/finishwrite_gk"
	"github.com/mit-pdos/gokv/tutorial/objectstore/dir/recordchunk_gk"
)

type WriteID = uint64

const (
	PrepareWriteId uint64 = 1
	RecordChunkId  uint64 = 2
	FinishWriteId  uint64 = 3
	PrepareReadId  uint64 = 4
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
	empty := make([]byte, 0)
	reply := new([]byte)
	ck.client.Call(PrepareWriteId, empty, reply, 100 /*ms*/)
	return ParsePreparedWrite(*reply)
}

// From chunk
func (ck *Clerk) RecordChunk(args recordchunk_gk.S) {
	req := recordchunk_gk.Marshal(make([]byte, 0), args)
	reply := new([]byte)
	ck.client.Call(RecordChunkId, req, reply, 100 /*ms*/)
}

// From chunk
func (ck *Clerk) FinishWrite(args finishwrite_gk.S) {
	req := finishwrite_gk.Marshal(make([]byte, 0), args)
	reply := new([]byte)
	ck.client.Call(FinishWriteId, req, reply, 100 /*ms*/)
}

func (ck *Clerk) PrepareRead(keyname string) PreparedRead {
	req := []byte(keyname)
	reply := new([]byte)
	ck.client.Call(PrepareReadId, req, reply, 100 /*ms*/)
	return ParsePreparedRead(*reply)
}
