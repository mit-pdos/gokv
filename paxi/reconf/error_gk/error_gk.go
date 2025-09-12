package error_gk

import "github.com/tchajed/marshal"

type E uint32

const (
	ENone         E = 0
	ETermStale    E = 1
	ENotLeader    E = 2
	EQuorumFailed E = 3
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
	1: "ETermStale",
	2: "ENotLeader",
	3: "EQuorumFailed",
}

var Value = map[string]uint32{
	"ENone":         0,
	"ETermStale":    1,
	"ENotLeader":    2,
	"EQuorumFailed": 3,
}

func (e E) String() string {
	return Name[uint32(e)]
}
