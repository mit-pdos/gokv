package error_gk

import "github.com/tchajed/marshal"

type E uint32

const (
	ENone       E = 0
	EEpochStale E = 1
	EOutOfOrder E = 2
	ETimeout    E = 3
	ENotLeader  E = 4
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
	1: "EEpochStale",
	2: "EOutOfOrder",
	3: "ETimeout",
	4: "ENotLeader",
}

var Value = map[string]uint32{
	"ENone":       0,
	"EEpochStale": 1,
	"EOutOfOrder": 2,
	"ETimeout":    3,
	"ENotLeader":  4,
}

func (e E) String() string {
	return Name[uint32(e)]
}
