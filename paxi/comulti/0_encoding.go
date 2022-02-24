package comulti

import (
	"github.com/tchajed/marshal"
)

func encodeBool(a bool) []byte {
	e := marshal.NewEnc(1)
	e.PutBool(a)
	return e.Finish()
}

func decodeBool(raw []byte) bool {
	return marshal.NewDec(raw).GetBool()
}

func encodeUint64(a uint64) []byte {
	e := marshal.NewEnc(8)
	e.PutInt(a)
	return e.Finish()
}

func decodeUint64(raw []byte) uint64 {
	return marshal.NewDec(raw).GetInt()
}

type PrepareReply struct {
	Success bool
	Log     []Entry // full log;
	Pn      uint64
}

func encodePrepareReply(rep *PrepareReply) []byte {
	e := marshal.NewEnc(1 + 8 + 8*uint64(len(rep.Log)))
	e.PutBool(rep.Success)
	e.PutInt(rep.Pn)
	e.PutInts(rep.Log)
	return e.Finish()
}

func decodePrepareReply(rawRep []byte) *PrepareReply {
	rep := new(PrepareReply)
	d := marshal.NewDec(rawRep) // (1 + 8 + 8 * uint64(len(rep.Log)))
	rep.Success = d.GetBool()
	rep.Pn = d.GetInt()
	rep.Log = d.GetInts((uint64(len(rawRep)) - 9) / 8)
	return rep
}

type ProposeArgs struct {
	Pn          uint64
	CommitIndex uint64
	Log         []Entry
}

func encodeProposeArgs(args *ProposeArgs) []byte {
	e := marshal.NewEnc(8 + 8 + 8*uint64(len(args.Log)))
	e.PutInt(args.Pn)
	e.PutInt(args.CommitIndex)
	e.PutInts(args.Log)
	return e.Finish()
}

func decodeProposeArgs(rawArgs []byte) *ProposeArgs {
	args := new(ProposeArgs)
	d := marshal.NewDec(rawArgs) // (1 + 8 + 8 * uint64(len(rep.Log)))
	args.Pn = d.GetInt()
	args.CommitIndex = d.GetInt()
	args.Log = d.GetInts((uint64(len(rawArgs)) - 16) / 8)
	return args
}
