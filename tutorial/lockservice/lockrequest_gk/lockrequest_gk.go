//--------------------------------------
// This file is autogenerated by grackle
// DO NOT MANUALLY EDIT THIS FILE
//--------------------------------------

package lockrequest_gk

import (
	"github.com/tchajed/marshal"
)

type S struct {
	Id uint64
}

func Marshal(prefix []byte, l S) []byte {
	var enc = prefix

	enc = marshal.WriteInt(enc, l.Id)

	return enc
}

func Unmarshal(s []byte) (S, []byte) {
	var enc = s // Needed for goose compatibility
	var id uint64

	id, enc = marshal.ReadInt(enc)

	return S{
		Id: id,
	}, enc
}
