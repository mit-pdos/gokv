package chunkhandle_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	Addr        uint64
	ContentHash string
}

func (c *S) approxSize() uint64 {
	return 0
}

func Marshal(c *S, prefix []byte) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, c.Addr)
	contenthashBytes := []byte(c.ContentHash)
	enc = marshal.WriteInt(enc, uint64(len(contenthashBytes)))
	enc = marshal.WriteBytes(enc, contenthashBytes)

	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	c := new(S)
	var enc = s // Needed for goose compatibility

	c.Addr, enc = marshal.ReadInt(enc)
	var contentHashLen uint64
	var contentHashBytes []byte
	contentHashLen, enc = marshal.ReadInt(enc)
	contentHashBytes, enc = marshal.ReadBytes(enc, contentHashLen)
	c.ContentHash = string(contentHashBytes)

	return c, enc
}
