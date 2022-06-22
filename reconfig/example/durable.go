package example

import (
	pb "github.com/mit-pdos/gokv/reconfig/replica"
)

// No durability.

func Append(acceptedEpoch uint64, entry []byte) {
}

func SetLog(acceptedEpoch uint64, startIndex uint64, log []pb.LogEntry) {
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
