package example

import (
	pb "github.com/mit-pdos/gokv/reconfig/replica"
	"github.com/mit-pdos/gokv/reconfig/replica/logentry_gk"
)

// No durability.

func Append(entry logentry_gk.S) {
}

func SetLog(startIndex uint64, log []logentry_gk.S) {
}

func Truncate(index uint64) {
}

func SetEpoch(epoch uint64) {
}

func NonDurable() pb.DurableState {
	return pb.DurableState{
		Append:   Append,
		SetLog:   SetLog,
		Truncate: Truncate,
		SetEpoch: SetEpoch,
	}
}
