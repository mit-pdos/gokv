package mpaxos

import ()

type Error uint64

const (
	ENone       = Error(0)
	EEpochStale = Error(1)
	EOutOfOrder = Error(2)
	ETimeout    = Error(3)
	ENotLeader  = Error(4)
)

type applyAsFollowerArgs struct {
	epoch     uint64
	nextIndex uint64
	state     []byte
}

type applyAsFollowerReply struct {
	err Error
}

type enterNewEpochArgs struct {
	epoch uint64
}

type enterNewEpochReply struct {
	err           Error
	acceptedEpoch uint64
	nextIndex     uint64
	state         []byte
}

type applyReply struct {
	err Error
	ret []byte
}
