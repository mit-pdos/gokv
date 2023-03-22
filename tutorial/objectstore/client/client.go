package client

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/objectstore/chunk"
	"github.com/mit-pdos/gokv/tutorial/objectstore/dir"
)

type Clerk struct {
	dCk  *dir.Clerk
	chCk *chunk.ClerkPool
}

type Writer struct {
	writeId    dir.WriteID
	index      uint64
	keyname    string
	wg         *sync.WaitGroup
	ck         *Clerk
	chunkAddrs []grove_ffi.Address
}

func (ck *Clerk) PrepareWrite(keyname string) *Writer {
	w := ck.dCk.PrepareWrite()
	return &Writer{
		writeId:    w.Id,
		index:      0,
		keyname:    keyname,
		chunkAddrs: w.ChunkAddrs,
	}
}

func (w *Writer) AppendChunk(data []byte) {
	w.wg.Add(1)
	index := w.index
	w.index = w.index + 1
	go func() {
		addr := w.chunkAddrs[index%uint64(len(w.chunkAddrs))]
		args := chunk.WriteChunkArgs{
			WriteId: w.writeId,
			Chunk:   data,
			Index:   index,
		}
		w.ck.chCk.WriteChunk(addr, args)
		// XXX: do we want this? w.ck.dCK.RecordChunk(...)
		w.wg.Done()
	}()
}

func (w *Writer) Done() {
	w.wg.Wait()
	w.ck.dCk.FinishWrite(dir.FinishWriteArgs{
		WriteId: w.writeId,
		Keyname: w.keyname,
	})
}

type Reader struct {
	chunkHandles []dir.ChunkHandle
	index        uint64
	ck           *Clerk
}

func (ck *Clerk) PrepareRead(keyname string) *Reader {
	return &Reader{
		chunkHandles: ck.dCk.PrepareRead(keyname).Handles,
		index:        0,
	}
}

func (r *Reader) GetNextChunk() (bool, []byte) {
	if r.index >= uint64(len(r.chunkHandles)) {
		return false, nil
	}
	handle := r.chunkHandles[r.index]
	r.index += 1
	return true, r.ck.chCk.GetChunk(handle.Addr, handle.ContentHash)
}
