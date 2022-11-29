package kv

// Replicated KV server using simplelog for durability.
// This does not use a reply table for deduplication.

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/map_marshal"
	"github.com/mit-pdos/gokv/simplepb/simplelog"
	"github.com/tchajed/marshal"
)

type KVState struct {
	kvs map[string][]byte
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
type putArgs struct {
	key []byte
	val []byte
}

func encodePutArgs(args *putArgs) []byte {
	var enc = make([]byte, 1, 1+8+uint64(len(args.key))+uint64(len(args.val)))
	enc[0] = OP_PUT
	enc = marshal.WriteInt(enc, uint64(len(args.key)))
	enc = marshal.WriteBytes(enc, args.key)
	enc = marshal.WriteBytes(enc, args.val)
	return enc
}

func decodePutArgs(raw_args []byte) *putArgs {
	var enc = raw_args
	args := new(putArgs)

	var l uint64
	l, enc = marshal.ReadInt(enc)
	args.key = enc[:l]
	args.val = enc[l:]

	return args
}

type getArgs []byte

func encodeGetArgs(args getArgs) []byte {
	var enc = make([]byte, 1, 1+uint64(len(args)))
	enc[0] = OP_GET
	enc = marshal.WriteBytes(enc, args)
	return enc
}

func decodeGetArgs(raw_args []byte) getArgs {
	return raw_args
}

// end of marshalling

func (s *KVState) put(args *putArgs) []byte {
	s.kvs[string(args.key)] = args.val
	return make([]byte, 0)
}

func (s *KVState) get(args getArgs) []byte {
	return s.kvs[string(args)]
}

func (s *KVState) apply(args []byte) []byte {
	if args[0] == OP_PUT {
		return s.put(decodePutArgs(args[1:]))
	} else if args[0] == OP_GET {
		return s.get(decodeGetArgs(args[1:]))
	} // else if args[0] == OP_CONDITIONALPUT {
	//return s.Put(args[1:])
	//}
	panic("unexpected op type")
}

func (s *KVState) getState() []byte {
	return map_marshal.EncodeMapStringToBytes(s.kvs)
}

func (s *KVState) setState(snap []byte) {
	if len(snap) == 0 {
		s.kvs = make(map[string][]byte, 0)
	} else {
		s.kvs = map_marshal.DecodeMapStringToBytes(snap)
	}
}

func makeKVStateMachine() *simplelog.InMemoryStateMachine {
	s := new(KVState)
	s.kvs = make(map[string][]byte, 0)

	return &simplelog.InMemoryStateMachine{
		ApplyVolatile: s.apply,
		GetState:      s.getState,
		SetState:      s.setState,
	}
}

func Start(fname string, me grove_ffi.Address) {
	r := simplelog.MakePbServer(makeKVStateMachine(), fname)
	r.Serve(me)
}
