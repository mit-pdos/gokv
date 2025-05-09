//--------------------------------------
// This file is autogenerated by grackle
// DO NOT MANUALLY EDIT THIS FILE
//--------------------------------------

package get_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	OpId uint64
	Key  string
}

func Marshal(prefix []byte, g S) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, g.OpId)
	keyBytes := []byte(g.Key)
	enc = marshal.WriteInt(enc, uint64(len(keyBytes)))
	enc = marshal.WriteBytes(enc, keyBytes)

	return enc
}

func Unmarshal(s []byte) (S, []byte) {
	var enc = s // Needed for goose compatibility
	var opId uint64
	var key string

	opId, enc = marshal.ReadInt(enc)
	var keyLen uint64
	var keyBytes []byte
	keyLen, enc = marshal.ReadInt(enc)
	keyBytes, enc = marshal.ReadBytesCopy(enc, keyLen)
	key = string(keyBytes)

	return S{
		OpId: opId,
		Key:  key,
	}, enc
}
