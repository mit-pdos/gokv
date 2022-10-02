package st_reconfig

import (
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/tchajed/marshal"
)

type SetStateArgs struct {
	Epoch     uint64
	NextIndex uint64
	State     []byte
}

func EncodeSetStateArgs(args *SetStateArgs) []byte {
	var enc = make([]byte, 0, 8+uint64(len(args.State)))
	enc = marshal.WriteInt(enc, args.Epoch)
	enc = marshal.WriteInt(enc, args.NextIndex)
	enc = marshal.WriteBytes(enc, args.State)
	return enc
}

func DecodeSetStateArgs(enc_args []byte) *SetStateArgs {
	var enc = enc_args
	args := new(SetStateArgs)
	args.Epoch, enc = marshal.ReadInt(enc)
	args.NextIndex, enc = marshal.ReadInt(enc)
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
	Err       e.Error
	NextIndex uint64
	State     []byte
}

func EncodeGetStateReply(reply *GetStateReply) []byte {
	var enc = make([]byte, 0, 8+len(reply.State))
	enc = marshal.WriteInt(enc, reply.Err)
	enc = marshal.WriteInt(enc, reply.NextIndex)
	enc = marshal.WriteBytes(enc, reply.State)
	return enc
}

func DecodeGetStateReply(enc_reply []byte) *GetStateReply {
	var enc = enc_reply
	reply := new(GetStateReply)
	reply.Err, enc = marshal.ReadInt(enc)
	reply.NextIndex, enc = marshal.ReadInt(enc)
	reply.State = enc
	return reply
}
