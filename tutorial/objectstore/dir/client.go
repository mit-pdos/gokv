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

type PreparedWrite struct {
	Id         WriteID
	ChunkAddrs []grove_ffi.Address
}

func (ck *Clerk) PrepareWrite() PreparedWrite {
	panic("impl")
}

// From chunk
func (ck *Clerk) RecordChunk(writeId WriteID, server grove_ffi.Address, content_hash string,
	index uint64) {
	panic("impl")
}

// From chunk
func (ck *Clerk) FinishWrite(writeId WriteID, keyname string) {
}

type ChunkHandle struct {
	Addr        grove_ffi.Address
	ContentHash string
}

func (ck *Clerk) PrepareRead(keyname string) []ChunkHandle {
	panic("impl")
}
