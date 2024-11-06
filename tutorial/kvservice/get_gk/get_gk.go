package get_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	OpId uint64
	Key  string
}

func (g *S) approxSize() uint64 {
	return 0
}

func Marshal(g *S, prefix []byte) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, g.OpId)
	keyBytes := []byte(g.Key)
	enc = marshal.WriteInt(enc, uint64(len(keyBytes)))
	enc = marshal.WriteBytes(enc, keyBytes)

	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	g := new(S)
	var enc = s // Needed for goose compatibility

	g.OpId, enc = marshal.ReadInt(enc)
	var keyLen uint64
	var keyBytes []byte
	keyLen, enc = marshal.ReadInt(enc)
	keyBytes, enc = marshal.ReadBytes(enc, keyLen)
	g.Key = string(keyBytes)

	return g, enc
}
