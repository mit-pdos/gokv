package error_gk

import "github.com/tchajed/marshal"

type E uint32

const (
	ENone  E = 0
	EStale E = 1
)

func Marshal(enc []byte, e E) []byte {
	return marshal.WriteInt32(enc, uint32(e))
}

func Unmarshal(s []byte) (E, []byte) {
	e_raw, s := marshal.ReadInt32(s)
	return E(e_raw), s
}

var Name = map[uint32]string{
	0: "ENone",
	1: "EStale",
}

var Value = map[string]uint32{
	"ENone":  0,
	"EStale": 1,
}

func (e E) String() string {
	return Name[uint32(e)]
}
