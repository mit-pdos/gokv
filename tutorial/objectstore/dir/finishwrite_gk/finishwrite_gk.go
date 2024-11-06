package finishwrite_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	WriteId uint64
	Keyname string
}

func (f *S) approxSize() uint64 {
	return 0
}

func Marshal(f *S, prefix []byte) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, f.WriteId)
	keynameBytes := []byte(f.Keyname)
	enc = marshal.WriteInt(enc, uint64(len(keynameBytes)))
	enc = marshal.WriteBytes(enc, keynameBytes)

	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	f := new(S)
	var enc = s // Needed for goose compatibility

	f.WriteId, enc = marshal.ReadInt(enc)
	var keynameLen uint64
	var keynameBytes []byte
	keynameLen, enc = marshal.ReadInt(enc)
	keynameBytes, enc = marshal.ReadBytes(enc, keynameLen)
	f.Keyname = string(keynameBytes)

	return f, enc
}
