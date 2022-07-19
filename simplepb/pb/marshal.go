package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

type ApplyArgs struct {
	epoch uint64
	index uint64
	op    []byte
}

func EncodeApplyArgs(args *ApplyArgs) []byte {
	var enc = make([]byte, 0, 8 + 8 + uint64(len(args.op)))
	enc = marshal.WriteInt(enc, args.epoch)
	enc = marshal.WriteInt(enc, args.index)
	enc = marshal.WriteBytes(enc, args.op)
	return enc
}

func DecodeApplyArgs(enc []byte) *ApplyArgs {
	args := new(ApplyArgs)
	args.epoch, enc = marshal.ReadInt(enc)
	args.index, enc = marshal.ReadInt(enc)
	args.op = enc
	return args
}

type SetStateArgs struct {
	epoch uint64
	state []byte
}

func EncodeSetStateArgs(args *SetStateArgs) []byte {
	var enc = make([]byte, 0, 8 + uint64(len(args.state)))
	enc = marshal.WriteInt(enc, args.epoch)
	enc = marshal.WriteBytes(enc, args.state)
	return enc
}

func DecodeSetStateArgs(enc []byte) *SetStateArgs {
	args := new(SetStateArgs)
	args.epoch, enc = marshal.ReadInt(enc)
	args.state = enc
	return args
}

type GetStateArgs struct {
	epoch uint64
}

func EncodeGetStateArgs(args *GetStateArgs) []byte {
	var enc = make([]byte, 0, 8)
	enc = marshal.WriteInt(enc, args.epoch)
	return enc
}

func DecodeGetStateArgs(enc []byte) *GetStateArgs {
	args := new(GetStateArgs)
	args.epoch, enc = marshal.ReadInt(enc)
	return args
}

type GetStateReply struct {
	err   Error
	state []byte
}

func EncodeGetStateReply(reply *GetStateReply) []byte {
	var enc = make([]byte, 0, 8)
	enc = marshal.WriteInt(enc, reply.err)
	enc = marshal.WriteBytes(enc, reply.state)
	return enc
}

type BecomePrimaryArgs struct {
	epoch   uint64
	replicas []grove_ffi.Address
}

func EncodeBecomePrimaryArgs(args *BecomePrimaryArgs) []byte {
	var enc = make([]byte, 0, 8 + 8 + 8 * uint64(len(args.replicas)))
	enc = marshal.WriteInt(enc, args.epoch)
	enc = marshal.WriteInt(enc, uint64(len(args.replicas)))
	for _, h := range args.replicas {
		enc = marshal.WriteInt(enc, h)
	}
	return enc
}

func DecodeBecomePrimaryArgs(enc []byte) *BecomePrimaryArgs {
	args := new(BecomePrimaryArgs)
	args.epoch, enc = marshal.ReadInt(enc)
	replicasLen, enc := marshal.ReadInt(enc)
	args.replicas = make([]grove_ffi.Address, replicasLen)
	for i, _ := range args.replicas {
		args.replicas[i], enc = marshal.ReadInt(enc)
	}
	return args
}

func EncodeError(err Error) []byte {
	return marshal.WriteInt(make([]byte, 0, 8), err)
}

func DecodeError(enc []byte) Error {
	err, _ := marshal.ReadInt(enc)
	return err
}
