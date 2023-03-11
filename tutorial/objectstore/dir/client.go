package dir

import "github.com/mit-pdos/gokv/grove_ffi"

type Clerk struct {
}

func (ck *Clerk) PrepareWrite() WriteID {
	panic("impl")
}

// From chunk
func (ck *Clerk) RecordChunk(writeId WriteID, server grove_ffi.Address, content_hash uint64,
	index uint64) {
	panic("impl")
}

// From chunk
func (ck *Clerk) FinishWrite(writeId WriteID, keyname string) {
}

type ChunkHandle struct {
	Addr grove_ffi.Address
	ContentHash uint64
}

func (ck *Clerk) PrepareRead(keyname string) []ChunkHandle {
	panic("impl")
}
