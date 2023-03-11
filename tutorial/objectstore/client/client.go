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
	return &Writer{
		writeId:    ck.dCk.PrepareWrite(),
		index:      0,
		keyname:    keyname,
		chunkAddrs: nil, // FIXME: get from dir
	}
}

func (w *Writer) AppendChunk(chunk []byte) {
	w.wg.Add(1)
	index := w.index
	w.index = w.index + 1
	go func() {
		w.ck.chCk.WriteChunk(w.chunkAddrs[index%uint64(len(w.chunkAddrs))], w.writeId, chunk, index)
		// XXX: do we want this? w.ck.dCK.RecordChunk(...)
		w.wg.Done()
	}()
}

func (w *Writer) Done() {
	w.wg.Wait()
	w.ck.dCk.FinishWrite(w.writeId, w.keyname)
}

type Reader struct {
	chunkHandles []dir.ChunkHandle
	index        uint64
	ck           *Clerk
}

func (ck *Clerk) PrepareRead(keyname string) *Reader {
	return &Reader{
		chunkHandles: ck.dCk.PrepareRead(keyname),
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
