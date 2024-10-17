package lockrequest_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	Id uint64
}

func (l *S) approxSize() uint64 {
	return 0
}

func Marshal(l *S, prefix []byte) []byte {
	var enc = prefix
	enc = marshal.WriteInt(enc, l.id)
	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	l := new(S)
	var enc = s // Needed for goose compatibility
	l.id, enc = marshal.ReadInt(enc)
	return l, enc
}
