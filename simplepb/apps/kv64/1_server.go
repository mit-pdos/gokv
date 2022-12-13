package kv64

// Replicated KV server using simplelog for durability.
// This does not use a reply table for deduplication.

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/map_marshal"
	"github.com/mit-pdos/gokv/simplepb/simplelog"
	"github.com/tchajed/marshal"
)

type KVState struct {
	kvs map[uint64][]byte
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

func (s *KVState) apply(args []byte) []byte {
	var ret []byte
	n := len(args)
	if args[0] == OP_PUT {
		ret = s.put(DecodePutArgs(args[1:n]))
	} else if args[0] == OP_GET {
		ret = s.get(DecodeGetArgs(args[1:n]))
	} else {
		panic("unexpected op type")
	}
	return ret
}

func (s *KVState) getState() []byte {
	return map_marshal.EncodeMapU64ToBytes(s.kvs)
}

func (s *KVState) setState(snap []byte) {
	s.kvs, _ = map_marshal.DecodeMapU64ToBytes(snap)
}

func MakeKVStateMachine() *simplelog.InMemoryStateMachine {
	s := new(KVState)
	s.kvs = make(map[uint64][]byte, 0)

	return &simplelog.InMemoryStateMachine{
		ApplyVolatile: s.apply,
		GetState:      func() []byte { return s.getState() },
		SetState:      s.setState,
	}
}

func Start(fname string, me grove_ffi.Address) {
	r := simplelog.MakePbServer(MakeKVStateMachine(), fname)
	r.Serve(me)
}