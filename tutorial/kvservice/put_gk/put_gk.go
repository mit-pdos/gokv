package put_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	OpId uint64
	Key  string
	Val  string
}

func (p *S) approxSize() uint64 {
	return 0
}

func Marshal(p *S, prefix []byte) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, p.OpId)
	keyBytes := []byte(p.Key)
	enc = marshal.WriteInt(enc, uint64(len(keyBytes)))
	enc = marshal.WriteBytes(enc, keyBytes)
	valBytes := []byte(p.Val)
	enc = marshal.WriteInt(enc, uint64(len(valBytes)))
	enc = marshal.WriteBytes(enc, valBytes)

	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	p := new(S)
	var enc = s // Needed for goose compatibility

	p.OpId, enc = marshal.ReadInt(enc)
	var keyLen uint64
	var keyBytes []byte
	keyLen, enc = marshal.ReadInt(enc)
	keyBytes, enc = marshal.ReadBytesCopy(enc, keyLen)
	p.Key = string(keyBytes)
	var valLen uint64
	var valBytes []byte
	valLen, enc = marshal.ReadInt(enc)
	valBytes, enc = marshal.ReadBytesCopy(enc, valLen)
	p.Val = string(valBytes)

	return p, enc
}
