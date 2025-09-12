package err_gk

import "github.com/tchajed/marshal"

type E uint32

const (
	None         E = 0
	Stale        E = 1
	OutOfOrder   E = 2
	Timeout      E = 3
	EmptyConfig  E = 4
	NotLeader    E = 5
	Sealed       E = 6
	LeaseExpired E = 7
	Leased       E = 8
)

func Marshal(enc []byte, e E) []byte {
	return marshal.WriteInt32(enc, uint32(e))
}

func Unmarshal(s []byte) (E, []byte) {
	e_raw, s := marshal.ReadInt32(s)
	return E(e_raw), s
}

var Name = map[uint32]string{
	0: "None",
	1: "Stale",
	2: "OutOfOrder",
	3: "Timeout",
	4: "EmptyConfig",
	5: "NotLeader",
	6: "Sealed",
	7: "LeaseExpired",
	8: "Leased",
}

var Value = map[string]uint32{
	"None":         0,
	"Stale":        1,
	"OutOfOrder":   2,
	"Timeout":      3,
	"EmptyConfig":  4,
	"NotLeader":    5,
	"Sealed":       6,
	"LeaseExpired": 7,
	"Leased":       8,
}

func (e E) String() string {
	return Name[uint32(e)]
}
