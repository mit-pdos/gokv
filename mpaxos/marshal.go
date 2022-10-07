package mpaxos

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

func decodeApplyAsFollowerArgs(s []byte) *applyAsFollowerArgs {
	panic("impl")
}

func encodeApplyAsFollowerArgs(o *applyAsFollowerArgs) []byte {
	panic("impl")
}

type applyAsFollowerReply struct {
	err Error
}

func decodeApplyAsFollowerReply(s []byte) *applyAsFollowerReply {
	panic("impl")
}

func encodeApplyAsFollowerReply(o *applyAsFollowerReply) []byte {
	panic("impl")
}

type enterNewEpochArgs struct {
	epoch uint64
}

func decodeEnterNewEpochArgs(s []byte) *enterNewEpochArgs {
	panic("impl")
}

func encodeEnterNewEpochArgs(o *enterNewEpochArgs) []byte {
	panic("impl")
}

type enterNewEpochReply struct {
	err           Error
	acceptedEpoch uint64
	nextIndex     uint64
	state         []byte
}

func decodeEnterNewEpochReply(s []byte) *enterNewEpochReply {
	panic("impl")
}

func encodeEnterNewEpochReply(o *enterNewEpochReply) []byte {
	panic("impl")
}

type applyReply struct {
	err Error
	ret []byte
}

func encodeApplyReply(o *applyReply) []byte {
	panic("impl")
}

func decodeApplyReply(s []byte) *applyReply {
	panic("impl")
}
