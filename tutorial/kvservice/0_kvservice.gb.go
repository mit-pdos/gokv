package lockservice

import (
	"github.com/tchajed/marshal"
)

// TODO: these are generic
func EncodeBool(a bool) []byte {
	if a {
		return append(make([]byte, 0), 1)
	} else {
		return append(make([]byte, 0), 0)
	}
}

func DecodeBool(x []byte) bool {
	return x[0] == 1
}

func EncodeUint64(a uint64) []byte {
	return marshal.WriteInt(make([]byte, 0), a)
}

func DecodeUint64(x []byte) uint64 {
	a, _ := marshal.ReadInt(x)
	return a
}

// Put
type putArgs struct {
	opId uint64
	key  string
	val  string
}

func encodePutArgs(a *putArgs) []byte {
	var e = make([]byte, 0)
	e = marshal.WriteInt(e, a.opId)
	e = marshal.WriteInt(e, uint64(len(a.key)))
	e = marshal.WriteBytes(e, []byte(a.key))
	e = marshal.WriteBytes(e, []byte(a.val))
	return e
}

func decodePutArgs(x []byte) *putArgs {
	var e = x
	a := new(putArgs)
	a.opId, e = marshal.ReadInt(e)

	keyLen, e := marshal.ReadInt(e) // FIXME: does this get translated correctly?
	keyBytes, valBytes := marshal.ReadBytes(e, keyLen)
	a.key = string(keyBytes)
	a.val = string(valBytes)
	return a
}

// ConditionalPut
type conditionalPutArgs struct {
	opId        uint64
	key         string
	expectedVal string
	newVal      string
}

func encodeConditionalPutArgs(a *conditionalPutArgs) []byte {
	var e = make([]byte, 0)
	e = marshal.WriteInt(e, a.opId)

	e = marshal.WriteInt(e, uint64(len(a.key)))
	e = marshal.WriteBytes(e, []byte(a.key))

	e = marshal.WriteInt(e, uint64(len(a.expectedVal)))
	e = marshal.WriteBytes(e, []byte(a.expectedVal))

	e = marshal.WriteBytes(e, []byte(a.newVal))
	e = marshal.WriteBytes(e, []byte(a.newVal))
	return e
}

func decodeConditionalPutArgs(x []byte) *conditionalPutArgs {
	var e = x
	a := new(conditionalPutArgs)
	a.opId, e = marshal.ReadInt(e)

	keyLen, e := marshal.ReadInt(e)
	keyBytes, e := marshal.ReadBytes(e, keyLen)
	a.key = string(keyBytes)

	expectedValLen, e := marshal.ReadInt(e)
	expectedValBytes, newValBytes := marshal.ReadBytes(e, expectedValLen)

	a.expectedVal = string(expectedValBytes)
	a.newVal = string(newValBytes)
	return a
}

// Get
type getArgs struct {
	opId        uint64
	key         string
}

func encodeGetArgs(a *getArgs) []byte {
	var e = make([]byte, 0)
	e = marshal.WriteInt(e, a.opId)
	e = marshal.WriteBytes(e, []byte(a.key))
	return e
}

func decodeGetArgs(x []byte) *getArgs {
	var e = x
	var keyBytes []byte
	a := new(getArgs)
	a.opId, keyBytes = marshal.ReadInt(e)
	a.key = string(keyBytes)
	return a
}
