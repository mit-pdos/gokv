package reconf

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

func EncodeConfig(pre []byte, conf *Config) []byte {
	var enc []byte = pre

	enc = marshal.WriteInt(enc, uint64(len(conf.members)))
	enc = marshal.WriteInt(enc, uint64(len(conf.nextMembers)))
	for _, member := range conf.members {
		enc = marshal.WriteInt(enc, member)
	}
	for _, member := range conf.nextMembers {
		enc = marshal.WriteInt(enc, member)
	}

	return enc
}

func DecodeConfig(encoded []byte) *Config {
	var dec []byte = encoded
	conf := new(Config)
	numMembers, dec := marshal.ReadInt(dec)
	numNextMembers, dec := marshal.ReadInt(dec)

	conf.members = make([]grove_ffi.Address, numMembers)
	conf.nextMembers = make([]grove_ffi.Address, numNextMembers)

	for i := range conf.members {
		conf.members[i], dec = marshal.ReadInt(dec)
	}
	for i := range conf.nextMembers {
		conf.members[i], dec = marshal.ReadInt(dec)
	}
	return conf
}

type MonotonicValue struct {
	version uint64
	val     []byte
	conf    *Config
}

func EncodeMonotonicValue(pre []byte, mval *MonotonicValue) []byte {
	var enc []byte = pre
	enc = marshal.WriteInt(enc, mval.version)
	enc = marshal.WriteInt(enc, uint64(len(mval.val)))
	enc = marshal.WriteBytes(enc, mval.val)
	enc = EncodeConfig(enc, mval.conf)
	return enc
}

func DecodeMonotonicValue(encoded []byte) *MonotonicValue {
	mval := new(MonotonicValue)
	var dec []byte = encoded
	mval.version, dec = marshal.ReadInt(dec)
	valLen, dec := marshal.ReadInt(dec)
	mval.val, dec = dec[:valLen], dec[valLen:]
	mval.conf = DecodeConfig(dec)
	return mval
}

type PrepareReply struct {
	Success bool
	Term    uint64
	Val     *MonotonicValue
}

type ProposeArgs struct {
	Term uint64
	Val  *MonotonicValue
}

func EncProposeArgs(args *ProposeArgs) []byte {
	var enc []byte = make([]byte, 0)
	enc = marshal.WriteInt(enc, args.Term)
	enc = EncodeMonotonicValue(enc, args.Val)
	return enc
}
