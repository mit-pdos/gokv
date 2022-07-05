package example

import (
	pb "github.com/mit-pdos/gokv/reconfig/replica"
)

// No durability.

func Append(entry []byte) {
}

func SetLog(startIndex uint64, log []pb.LogEntry) {
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
