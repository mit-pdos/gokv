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

func encodeApplyAsFollowerArgs(o *applyAsFollowerArgs) []byte {
	var enc = make([]byte, 0, 8+8+len(o.state))
	enc = marshal.WriteInt(enc, o.epoch)
	enc = marshal.WriteInt(enc, o.nextIndex)
	enc = marshal.WriteBytes(enc, o.state)
	return enc
}

func decodeApplyAsFollowerArgs(s []byte) *applyAsFollowerArgs {
	var enc = s
	o := new(applyAsFollowerArgs)
	o.epoch, enc = marshal.ReadInt(enc)
	o.nextIndex, enc = marshal.ReadInt(enc)
	o.state = enc
	return o
}

type applyAsFollowerReply struct {
	err Error
}

func decodeApplyAsFollowerReply(s []byte) *applyAsFollowerReply {
	o := new(applyAsFollowerReply)
	err, _ := marshal.ReadInt(s)
	o.err = Error(err)
	return o
}

func encodeApplyAsFollowerReply(o *applyAsFollowerReply) []byte {
	var enc []byte = make([]byte, 0, 8)
	enc = marshal.WriteInt(enc, uint64(o.err))
	return enc
}

type enterNewEpochArgs struct {
	epoch uint64
}

func encodeEnterNewEpochArgs(o *enterNewEpochArgs) []byte {
	var enc []byte = make([]byte, 0, 8)
	enc = marshal.WriteInt(enc, o.epoch)
	return enc
}

func decodeEnterNewEpochArgs(s []byte) *enterNewEpochArgs {
	o := new(enterNewEpochArgs)
	o.epoch, _ = marshal.ReadInt(s)
	return o
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
	var enc = make([]byte, 0, 8+8+8+uint64(len(o.state)))
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
	var enc = make([]byte, 0, 8+uint64(len(o.ret)))
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

func encodePaxosState(ps *paxosState) []byte {
	panic("impl")
}

func decodePaxosState([]byte) *paxosState {
	panic("impl")
}
