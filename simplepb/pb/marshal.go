package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/tchajed/marshal"
)

type Op = []byte

type ApplyArgs struct {
	epoch uint64
	index uint64
	op    []byte
}

func EncodeApplyArgs(args *ApplyArgs) []byte {
	var enc = make([]byte, 0, 8+8+uint64(len(args.op)))
	enc = marshal.WriteInt(enc, args.epoch)
	enc = marshal.WriteInt(enc, args.index)
	enc = marshal.WriteBytes(enc, args.op)
	return enc
}

func DecodeApplyArgs(enc_args []byte) *ApplyArgs {
	var enc = enc_args
	args := new(ApplyArgs)
	args.epoch, enc = marshal.ReadInt(enc)
	args.index, enc = marshal.ReadInt(enc)
	args.op = enc
	return args
}

type SetStateArgs struct {
	Epoch uint64
	State []byte
}

func EncodeSetStateArgs(args *SetStateArgs) []byte {
	var enc = make([]byte, 0, 8+uint64(len(args.State)))
	enc = marshal.WriteInt(enc, args.Epoch)
	enc = marshal.WriteBytes(enc, args.State)
	return enc
}

func DecodeSetStateArgs(enc_args []byte) *SetStateArgs {
	var enc = enc_args
	args := new(SetStateArgs)
	args.Epoch, enc = marshal.ReadInt(enc)
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
	Err   e.Error
	State []byte
}

func EncodeGetStateReply(reply *GetStateReply) []byte {
	var enc = make([]byte, 0, 8)
	enc = marshal.WriteInt(enc, reply.Err)
	enc = marshal.WriteBytes(enc, reply.State)
	return enc
}

func DecodeGetStateReply(enc_reply []byte) *GetStateReply {
	var enc = enc_reply
	reply := new(GetStateReply)
	reply.Err, enc = marshal.ReadInt(enc)
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
	replicasLen, enc := marshal.ReadInt(enc)
	args.Replicas = make([]grove_ffi.Address, replicasLen)
	for i := range args.Replicas {
		args.Replicas[i], enc = marshal.ReadInt(enc)
	}
	return args
}
