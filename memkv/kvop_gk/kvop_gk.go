package kvop_gk

import "github.com/tchajed/marshal"

type E uint32

const (
	KV_FRESHCID        E = 0
	KV_PUT             E = 1
	KV_GET             E = 2
	KV_CONDITIONAL_PUT E = 3
	KV_INS_SHARD       E = 4
	KV_MOV_SHARD       E = 5
)

func Marshal(enc []byte, e E) []byte {
	return marshal.WriteInt32(enc, uint32(e))
}

func Unmarshal(s []byte) (E, []byte) {
	e_raw, s := marshal.ReadInt32(s)
	return E(e_raw), s
}

var Name = map[uint32]string{
	0: "KV_FRESHCID",
	1: "KV_PUT",
	2: "KV_GET",
	3: "KV_CONDITIONAL_PUT",
	4: "KV_INS_SHARD",
	5: "KV_MOV_SHARD",
}

var Value = map[string]uint32{
	"KV_FRESHCID":        0,
	"KV_PUT":             1,
	"KV_GET":             2,
	"KV_CONDITIONAL_PUT": 3,
	"KV_INS_SHARD":       4,
	"KV_MOV_SHARD":       5,
}

func (k E) String() string {
	return Name[uint32(k)]
}
