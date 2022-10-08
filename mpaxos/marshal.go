package mpaxos

import "github.com/tchajed/marshal"

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
	o := new(enterNewEpochReply)
	var enc = s

	var err uint64
	err, enc = marshal.ReadInt(enc)
	o.err = Error(err)

	o.acceptedEpoch, enc = marshal.ReadInt(enc)
	o.nextIndex, enc = marshal.ReadInt(enc)
	o.state = enc
	return o
}

func encodeEnterNewEpochReply(o *enterNewEpochReply) []byte {
	var enc = make([]byte, 0, 8 + 8 + 8 + uint64(len(o.state)))
	enc = marshal.WriteInt(enc, uint64(o.err))
	enc = marshal.WriteInt(enc, o.acceptedEpoch)
	enc = marshal.WriteInt(enc, o.nextIndex)
	enc = marshal.WriteBytes(enc, o.state)
	return enc
}

type applyReply struct {
	err Error
	ret []byte
}

func encodeApplyReply(o *applyReply) []byte {
	var enc = make([]byte, 0, 8 + uint64(len(o.ret)))
	enc = marshal.WriteInt(enc, uint64(o.err))
	enc = marshal.WriteBytes(enc, o.ret)
	return enc
}

func decodeApplyReply(s []byte) *applyReply {
	var enc = s
	o := new(applyReply)

	var err uint64
	err, enc = marshal.ReadInt(enc)
	o.err = Error(err)

	o.ret = enc
	return o
}
