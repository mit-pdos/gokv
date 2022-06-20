package reconf

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

func EncConfig(pre []byte, conf *Config) []byte {
	var enc []byte = pre

	enc = marshal.WriteInt(enc, uint64(len(conf.Members)))
	enc = marshal.WriteInt(enc, uint64(len(conf.NextMembers)))
	for _, member := range conf.Members {
		enc = marshal.WriteInt(enc, member)
	}
	for _, member := range conf.NextMembers {
		enc = marshal.WriteInt(enc, member)
	}

	return enc
}

func DecConfig(encoded []byte) (*Config, []byte) {
	var dec []byte = encoded

	conf := new(Config)

	numMembers, dec := marshal.ReadInt(dec)
	numNextMembers, dec := marshal.ReadInt(dec)

	conf.Members = make([]grove_ffi.Address, numMembers)
	conf.NextMembers = make([]grove_ffi.Address, numNextMembers)

	for i := range conf.Members {
		conf.Members[i], dec = marshal.ReadInt(dec)
	}

	for i := range conf.NextMembers {
		conf.Members[i], dec = marshal.ReadInt(dec)
	}
	return conf, dec
}

type MonotonicValue struct {
	version uint64
	val     []byte
	conf    *Config
}

func EncMonotonicValue(pre []byte, mval *MonotonicValue) []byte {
	var enc []byte = pre
	enc = marshal.WriteInt(enc, mval.version)
	enc = marshal.WriteInt(enc, uint64(len(mval.val)))
	enc = marshal.WriteBytes(enc, mval.val)
	enc = EncConfig(enc, mval.conf)
	return enc
}

func DecMonotonicValue(encoded []byte) (*MonotonicValue, []byte) {
	mval := new(MonotonicValue)
	var dec []byte = encoded
	mval.version, dec = marshal.ReadInt(dec)
	valLen, dec := marshal.ReadInt(dec)
	mval.val = dec[:valLen]
	dec = dec[valLen:]
	mval.conf, dec = DecConfig(dec)
	return mval, dec
}

type PrepareReply struct {
	Err  uint64
	Term uint64
	Val  *MonotonicValue
}

func EncPrepareReply(pre []byte, reply *PrepareReply) []byte {
	var enc []byte = pre
	enc = marshal.WriteInt(enc, reply.Err)
	enc = marshal.WriteInt(enc, reply.Term)
	enc = EncMonotonicValue(enc, reply.Val)

	return enc
}

func DecPrepareReply(encoded []byte) *PrepareReply {
	var dec []byte = encoded
	reply := new(PrepareReply)
	reply.Err, dec = marshal.ReadInt(dec)
	reply.Term, dec = marshal.ReadInt(dec)
	reply.Val, dec = DecMonotonicValue(dec)
	return reply
}

type ProposeArgs struct {
	Term uint64
	Val  *MonotonicValue
}

func EncProposeArgs(args *ProposeArgs) []byte {
	var enc []byte = make([]byte, 0)
	enc = marshal.WriteInt(enc, args.Term)
	enc = EncMonotonicValue(enc, args.Val)
	return enc
}

func DecProposeArgs(encoded []byte) (*ProposeArgs, []byte) {
	var dec []byte = encoded
	args := new(ProposeArgs)
	args.Term, dec = marshal.ReadInt(dec)
	args.Val, dec = DecMonotonicValue(dec)
	return args, dec
}

type TryCommitReply struct {
	err     uint64
	version uint64
}

func EncMembers(members []grove_ffi.Address) []byte {
	var enc []byte = make([]byte, 0)
	enc = marshal.WriteInt(enc, uint64(len(members)))
	for _, member := range members {
		enc = marshal.WriteInt(enc, member)
	}
	return enc
}

func DecMembers(encoded []byte) ([]grove_ffi.Address, []byte) {
	var dec []byte = encoded
	numMembers, dec := marshal.ReadInt(dec)
	members := make([]grove_ffi.Address, numMembers)
	for i := range members {
		members[i], dec = marshal.ReadInt(dec)
	}
	return members, dec
}
