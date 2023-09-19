package replica

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/vrsm/e"
	"github.com/tchajed/marshal"
)

type Op = []byte

type ApplyAsBackupArgs struct {
	epoch uint64
	index uint64
	op    []byte
}

func EncodeApplyAsBackupArgs(args *ApplyAsBackupArgs) []byte {
	var enc = make([]byte, 0, 8+8+uint64(len(args.op)))
	enc = marshal.WriteInt(enc, args.epoch)
	enc = marshal.WriteInt(enc, args.index)
	enc = marshal.WriteBytes(enc, args.op)
	return enc
}

func DecodeApplyAsBackupArgs(enc_args []byte) *ApplyAsBackupArgs {
	var enc = enc_args
	args := new(ApplyAsBackupArgs)
	args.epoch, enc = marshal.ReadInt(enc)
	args.index, enc = marshal.ReadInt(enc)
	args.op = enc
	return args
}

type SetStateArgs struct {
	Epoch              uint64
	NextIndex          uint64
	CommittedNextIndex uint64
	State              []byte
}

func EncodeSetStateArgs(args *SetStateArgs) []byte {
	var enc = make([]byte, 0, 8+uint64(len(args.State)))
	enc = marshal.WriteInt(enc, args.Epoch)
	enc = marshal.WriteInt(enc, args.NextIndex)
	enc = marshal.WriteInt(enc, args.CommittedNextIndex)
	enc = marshal.WriteBytes(enc, args.State)
	return enc
}

func DecodeSetStateArgs(enc_args []byte) *SetStateArgs {
	var enc = enc_args
	args := new(SetStateArgs)
	args.Epoch, enc = marshal.ReadInt(enc)
	args.NextIndex, enc = marshal.ReadInt(enc)
	args.CommittedNextIndex, enc = marshal.ReadInt(enc)
	args.State = enc
	return args
}

type GetStateArgs struct {
	Epoch uint64
}

func EncodeGetStateArgs(args *GetStateArgs) []byte {
	var enc = make([]byte, 0, 8)
	enc = marshal.WriteInt(enc, args.Epoch)
	return enc
}

func DecodeGetStateArgs(enc []byte) *GetStateArgs {
	args := new(GetStateArgs)
	args.Epoch, _ = marshal.ReadInt(enc)
	return args
}

type GetStateReply struct {
	Err                e.Error
	NextIndex          uint64
	CommittedNextIndex uint64
	State              []byte
}

func EncodeGetStateReply(reply *GetStateReply) []byte {
	var enc = make([]byte, 0, 8+len(reply.State))
	enc = marshal.WriteInt(enc, reply.Err)
	enc = marshal.WriteInt(enc, reply.NextIndex)
	enc = marshal.WriteInt(enc, reply.CommittedNextIndex)
	enc = marshal.WriteBytes(enc, reply.State)
	return enc
}

func DecodeGetStateReply(enc_reply []byte) *GetStateReply {
	var enc = enc_reply
	reply := new(GetStateReply)
	reply.Err, enc = marshal.ReadInt(enc)
	reply.NextIndex, enc = marshal.ReadInt(enc)
	reply.CommittedNextIndex, enc = marshal.ReadInt(enc)
	reply.State = enc
	return reply
}

type BecomePrimaryArgs struct {
	Epoch    uint64
	Replicas []grove_ffi.Address
}

func EncodeBecomePrimaryArgs(args *BecomePrimaryArgs) []byte {
	var enc = make([]byte, 0, 8+8+8*uint64(len(args.Replicas)))
	enc = marshal.WriteInt(enc, args.Epoch)
	enc = marshal.WriteInt(enc, uint64(len(args.Replicas)))
	for _, h := range args.Replicas {
		enc = marshal.WriteInt(enc, h)
	}
	return enc
}

func DecodeBecomePrimaryArgs(enc_args []byte) *BecomePrimaryArgs {
	var enc = enc_args
	args := new(BecomePrimaryArgs)
	args.Epoch, enc = marshal.ReadInt(enc)
	var replicasLen uint64
	replicasLen, enc = marshal.ReadInt(enc)
	args.Replicas = make([]grove_ffi.Address, replicasLen)
	for i := range args.Replicas {
		args.Replicas[i], enc = marshal.ReadInt(enc)
	}
	return args
}

type ApplyReply struct {
	Err   e.Error
	Reply []byte
}

func EncodeApplyReply(reply *ApplyReply) []byte {
	var enc = make([]byte, 0, 8+uint64(len(reply.Reply)))
	enc = marshal.WriteInt(enc, reply.Err)
	enc = marshal.WriteBytes(enc, reply.Reply)
	return enc
}

func DecodeApplyReply(enc_reply []byte) *ApplyReply {
	var enc = enc_reply
	reply := new(ApplyReply)
	reply.Err, enc = marshal.ReadInt(enc)
	reply.Reply = enc // XXX: re-slices the enc_reply, so there 8 bytes in front
	// that will sit around until ApplyReply is deallocated
	return reply
}

type IncreaseCommitArgs = uint64

func EncodeIncreaseCommitArgs(args IncreaseCommitArgs) []byte {
	return marshal.WriteInt(nil, args)
}

func DecodeIncreaseCommitArgs(args []byte) IncreaseCommitArgs {
	a, _ := marshal.ReadInt(args)
	return a
}
