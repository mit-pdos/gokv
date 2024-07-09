package paxos

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
	enc := make([]byte, 0, 8+8+len(o.state))
	enc = marshal.WriteInt(enc, o.epoch)
	enc = marshal.WriteInt(enc, o.nextIndex)
	enc = marshal.WriteBytes(enc, o.state)
	return enc
}

func decodeApplyAsFollowerArgs(enc []byte) *applyAsFollowerArgs {
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
	o := &applyAsFollowerReply{}
	err, _ := marshal.ReadInt(s)
	o.err = Error(err)
	return o
}

func encodeApplyAsFollowerReply(o *applyAsFollowerReply) []byte {
	enc := make([]byte, 0, 8)
	enc = marshal.WriteInt(enc, uint64(o.err))
	return enc
}

type enterNewEpochArgs struct {
	epoch uint64
}

func encodeEnterNewEpochArgs(o *enterNewEpochArgs) []byte {
	enc := make([]byte, 0, 8)
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

func decodeEnterNewEpochReply(enc []byte) *enterNewEpochReply {
	o := &enterNewEpochReply{}

	err, enc := marshal.ReadInt(enc)
	o.err = Error(err)

	o.acceptedEpoch, enc = marshal.ReadInt(enc)
	o.nextIndex, enc = marshal.ReadInt(enc)
	o.state = enc
	return o
}

func encodeEnterNewEpochReply(o *enterNewEpochReply) []byte {
	enc := make([]byte, 0, 8+8+8+uint64(len(o.state)))
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
	enc := make([]byte, 0, 8+uint64(len(o.ret)))
	enc = marshal.WriteInt(enc, uint64(o.err))
	enc = marshal.WriteBytes(enc, o.ret)
	return enc
}

func decodeApplyReply(enc []byte) *applyReply {
	o := &applyReply{}

	err, enc := marshal.ReadInt(enc)
	o.err = Error(err)

	o.ret = enc
	return o
}

func boolToU64(b bool) uint64 {
	if b {
		return 1
	} else {
		return 0
	}
}

func encodePaxosState(ps *paxosState) []byte {
	var e = make([]byte, 0)
	e = marshal.WriteInt(e, ps.epoch)
	e = marshal.WriteInt(e, ps.acceptedEpoch)
	e = marshal.WriteInt(e, ps.nextIndex)
	e = marshal.WriteInt(e, boolToU64(ps.isLeader))
	e = marshal.WriteBytes(e, ps.state)
	return e
}

func decodePaxosState(enc []byte) *paxosState {
	var e []byte = enc
	var leaderInt uint64
	ps := new(paxosState)
	ps.epoch, e = marshal.ReadInt(e)
	ps.acceptedEpoch, e = marshal.ReadInt(e)
	ps.nextIndex, e = marshal.ReadInt(e)
	leaderInt, ps.state = marshal.ReadInt(e)
	ps.isLeader = (leaderInt == 1)
	return ps
}
