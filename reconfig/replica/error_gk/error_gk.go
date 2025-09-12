package error_gk

import "github.com/tchajed/marshal"

type E uint32

const (
	ENone             E = 0
	ENotPrimary       E = 1
	EStale            E = 2
	EAppendOutOfOrder E = 3
	ETruncated        E = 4
	EIncompleteLog    E = 5
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
	1: "ENotPrimary",
	2: "EStale",
	3: "EAppendOutOfOrder",
	4: "ETruncated",
	5: "EIncompleteLog",
}

var Value = map[string]uint32{
	"ENone":             0,
	"ENotPrimary":       1,
	"EStale":            2,
	"EAppendOutOfOrder": 3,
	"ETruncated":        4,
	"EIncompleteLog":    5,
}

func (e E) String() string {
	return Name[uint32(e)]
}
