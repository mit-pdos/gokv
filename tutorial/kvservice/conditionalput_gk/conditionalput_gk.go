package conditionalput_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	OpId        uint64
	Key         string
	ExpectedVal string
	NewVal      string
}

func (c *S) approxSize() uint64 {
	return 0
}

func Marshal(c *S, prefix []byte) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, c.OpId)
	keyBytes := []byte(c.Key)
	enc = marshal.WriteInt(enc, uint64(len(keyBytes)))
	enc = marshal.WriteBytes(enc, keyBytes)
	expectedvalBytes := []byte(c.ExpectedVal)
	enc = marshal.WriteInt(enc, uint64(len(expectedvalBytes)))
	enc = marshal.WriteBytes(enc, expectedvalBytes)
	newvalBytes := []byte(c.NewVal)
	enc = marshal.WriteInt(enc, uint64(len(newvalBytes)))
	enc = marshal.WriteBytes(enc, newvalBytes)

	return enc
}

func Unmarshal(s []byte) (*S, []byte) {
	c := new(S)
	var enc = s // Needed for goose compatibility

	c.OpId, enc = marshal.ReadInt(enc)
	var keyLen uint64
	var keyBytes []byte
	keyLen, enc = marshal.ReadInt(enc)
	keyBytes, enc = marshal.ReadBytes(enc, keyLen)
	c.Key = string(keyBytes)
	var expectedValLen uint64
	var expectedValBytes []byte
	expectedValLen, enc = marshal.ReadInt(enc)
	expectedValBytes, enc = marshal.ReadBytes(enc, expectedValLen)
	c.ExpectedVal = string(expectedValBytes)
	var newValLen uint64
	var newValBytes []byte
	newValLen, enc = marshal.ReadInt(enc)
	newValBytes, enc = marshal.ReadBytes(enc, newValLen)
	c.NewVal = string(newValBytes)

	return c, enc
}
