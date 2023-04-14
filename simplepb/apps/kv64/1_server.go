package kv64

// Replicated KV server using simplelog for durability.
// This does not use a reply table for deduplication.

import (
	"github.com/mit-pdos/gokv/map_marshal"
	"github.com/mit-pdos/gokv/simplepb/apps/eesm"
	"github.com/tchajed/marshal"
)

type KVState struct {
	kvs     map[uint64][]byte
	minVnum uint64
	vnums   map[uint64]uint64
}

// Ops include:
// Put(k, v)
// Get(k)
// // ConditionalPut(k, v, expected_v)
const (
	OP_PUT = byte(0)
	OP_GET = byte(1)
	// OP_CONDITIONALPUT = byte(2)
)

// begin arg structs and marshalling
type PutArgs struct {
	Key uint64
	Val []byte
}

func EncodePutArgs(args *PutArgs) []byte {
	var enc = make([]byte, 1, 9)
	enc[0] = OP_PUT
	enc = marshal.WriteInt(enc, args.Key)
	enc = marshal.WriteBytes(enc, args.Val)
	return enc
}

func DecodePutArgs(raw_args []byte) *PutArgs {
	var enc = raw_args
	args := new(PutArgs)
	args.Key, args.Val = marshal.ReadInt(enc)
	return args
}

type getArgs = uint64

func EncodeGetArgs(args getArgs) []byte {
	var enc = make([]byte, 1, 9)
	enc[0] = OP_GET
	enc = marshal.WriteInt(enc, args)
	return enc
}

func DecodeGetArgs(raw_args []byte) getArgs {
	key, _ := marshal.ReadInt(raw_args)
	return key
}

// end of marshalling

func (s *KVState) put(args *PutArgs) []byte {
	s.kvs[args.Key] = args.Val
	return make([]byte, 0)
}

func (s *KVState) get(args getArgs) []byte {
	return s.kvs[args]
}

func (s *KVState) apply(args []byte, vnum uint64) []byte {
	var ret []byte
	n := len(args)
	if args[0] == OP_PUT {
		args := DecodePutArgs(args[1:n])
		ret = s.put(args)
		s.vnums[args.Key] = vnum
	} else if args[0] == OP_GET {
		key := DecodeGetArgs(args[1:n])
		ret = s.get(key)
		s.vnums[key] = vnum
	} else {
		panic("unexpected op type")
	}
	return ret
}

func (s *KVState) getState() []byte {
	return map_marshal.EncodeMapU64ToBytes(s.kvs)
}

func (s *KVState) setState(snap []byte, vnum uint64) {
	s.kvs, _ = map_marshal.DecodeMapU64ToBytes(snap)
	s.vnums = make(map[uint64]uint64)
	s.minVnum = vnum
}

func (s *KVState) applyReadonly(args []byte) (uint64, []byte) {
	if args[0] == OP_PUT {
		panic("unexpectedly got put as readonly op")
	} else if args[0] != OP_GET {
		panic("unexpected op type")
	}
	n := len(args)
	key := DecodeGetArgs(args[1:n])
	ret := s.get(key)
	vnum, ok := s.vnums[key]
	if ok {
		return vnum, ret
	} else {
		return s.minVnum, ret
	}
}

func MakeKVStateMachine() *eesm.VersionedStateMachine {
	s := new(KVState)
	s.kvs = make(map[uint64][]byte, 0)

	return &eesm.VersionedStateMachine{
		ApplyVolatile: s.apply,
		ApplyReadonly: s.applyReadonly,
		GetState:      func() []byte { return s.getState() },
		SetState:      s.setState,
	}
}
