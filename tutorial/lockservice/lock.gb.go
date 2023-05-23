package lockservice

import "github.com/tchajed/marshal"

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
