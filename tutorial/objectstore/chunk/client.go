package chunk

import (
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/grove_ffi"
)

type WriteID = uint64

type ClerkPool struct {
	// TODO: connman for a bunch of chunk servers
	cm *connman.ConnMan
}

func (ck *ClerkPool) WriteChunk(addr grove_ffi.Address, args WriteChunkArgs) {
	req := MarshalWriteChunkArgs(args)
	reply := new([]byte)
	ck.cm.CallAtLeastOnce(addr, WriteChunkId, req, reply, 100 /*ms*/)
}

func (ck *ClerkPool) GetChunk(addr grove_ffi.Address, content_hash string) []byte {
	req := []byte(content_hash)
	reply := new([]byte)
	ck.cm.CallAtLeastOnce(addr, GetChunkId, req, reply, 100 /*ms*/)
	return *reply
}
