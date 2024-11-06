package writechunk_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	WriteId uint64
	Chunk   []byte
	Index   uint64
}

func (w *S) approxSize() uint64 {
	return 0
}

func Marshal(w *S, prefix []byte) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, w.WriteId)
	enc = marshal.WriteInt(enc, uint64(len(w.Chunk)))
	enc = marshal.WriteBytes(enc, w.Chunk)
	enc = marshal.WriteInt(enc, w.Index)

	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	w := new(S)
	var enc = s // Needed for goose compatibility

	w.WriteId, enc = marshal.ReadInt(enc)
	var chunkLen uint64
	var chunkBytes []byte
	chunkLen, enc = marshal.ReadInt(enc)
	chunkBytes, enc = marshal.ReadBytes(enc, chunkLen)
	w.Chunk = chunkBytes
	w.Index, enc = marshal.ReadInt(enc)

	return w, enc
}
