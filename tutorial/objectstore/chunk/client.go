package chunk

import "github.com/mit-pdos/gokv/grove_ffi"

type WriteID = uint64

type ClerkPool struct {
}

func (ck *ClerkPool) WriteChunk(addr grove_ffi.Address, writeId WriteID, chunk []byte, index uint64) {
	panic("impl")
}

func (ck *ClerkPool) GetChunk(addr grove_ffi.Address, content_hash uint64) []byte {
	panic("impl")
}
