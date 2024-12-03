package recordchunk_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	WriteId     uint64
	Server      uint64
	ContentHash string
	Index       uint64
}

func (r *S) approxSize() uint64 {
	return 0
}

func Marshal(r *S, prefix []byte) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, r.WriteId)
	enc = marshal.WriteInt(enc, r.Server)
	contenthashBytes := []byte(r.ContentHash)
	enc = marshal.WriteInt(enc, uint64(len(contenthashBytes)))
	enc = marshal.WriteBytes(enc, contenthashBytes)
	enc = marshal.WriteInt(enc, r.Index)

	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	r := new(S)
	var enc = s // Needed for goose compatibility

	r.WriteId, enc = marshal.ReadInt(enc)
	r.Server, enc = marshal.ReadInt(enc)
	var contentHashLen uint64
	var contentHashBytes []byte
	contentHashLen, enc = marshal.ReadInt(enc)
	contentHashBytes, enc = marshal.ReadBytesCopy(enc, contentHashLen)
	r.ContentHash = string(contentHashBytes)
	r.Index, enc = marshal.ReadInt(enc)

	return r, enc
}
