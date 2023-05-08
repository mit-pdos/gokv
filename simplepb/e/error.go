package e

import (
	"github.com/tchajed/marshal"
)

type Error = uint64

const (
	None         = uint64(0)
	Stale        = uint64(1)
	OutOfOrder   = uint64(2)
	Timeout      = uint64(3)
	EmptyConfig  = uint64(4)
	NotLeader    = uint64(5)
	Sealed       = uint64(6)
	LeaseExpired = uint64(7)
	Leased       = uint64(8)
)

func EncodeError(err Error) []byte {
	return marshal.WriteInt(make([]byte, 0, 8), err)
}

func DecodeError(enc []byte) Error {
	err, _ := marshal.ReadInt(enc)
	return err
}
