package ctr

import (
	"github.com/tchajed/marshal"
)

type PutArgs struct {
	cid uint64
	seq uint64
	v   uint64
}

func EncPutArgs(args *PutArgs) []byte {
	enc := marshal.NewEnc(uint64(24))
	enc.PutInt(args.cid)
	enc.PutInt(args.seq)
	enc.PutInt(args.v)
	return enc.Finish()
}

func DecPutArgs(raw_args []byte) *PutArgs {
	dec := marshal.NewDec(raw_args)
	args := new(PutArgs)
	args.cid = dec.GetInt()
	args.seq = dec.GetInt()
	args.v = dec.GetInt()
	return args
}
