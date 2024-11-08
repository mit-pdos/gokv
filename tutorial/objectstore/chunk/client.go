package chunk

import (
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/objectstore/chunk/writechunk_gk"
)

type WriteID = uint64

// rpc ids
const (
	WriteChunkId uint64 = 1
	GetChunkId   uint64 = 2
)

type ClerkPool struct {
	// TODO: connman for a bunch of chunk servers
	cm *connman.ConnMan
}

func (ck *ClerkPool) WriteChunk(addr grove_ffi.Address, args writechunk_gk.S) {
	req := writechunk_gk.Marshal(&args, make([]byte, 0))
	reply := new([]byte)
	ck.cm.CallAtLeastOnce(addr, WriteChunkId, req, reply, 100 /*ms*/)
}

func (ck *ClerkPool) GetChunk(addr grove_ffi.Address, content_hash string) []byte {
	req := []byte(content_hash)
	reply := new([]byte)
	ck.cm.CallAtLeastOnce(addr, GetChunkId, req, reply, 100 /*ms*/)
	return *reply
}
