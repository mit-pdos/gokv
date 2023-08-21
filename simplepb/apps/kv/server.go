package kv

// Replicated KV server using simplelog for durability.
// This does not use a reply table for deduplication.

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/map_string_marshal"
	"github.com/mit-pdos/gokv/simplepb/apps/esm"
	"github.com/mit-pdos/gokv/simplepb/simplelog"
	"github.com/tchajed/marshal"
)

type KVState struct {
	kvs     map[string]string
	vnums   map[string]uint64
	minVnum uint64
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
	Key string
	Val string
}

func EncodePutArgs(args *PutArgs) []byte {
	var enc = make([]byte, 1, 1+8+uint64(len(args.Key))+uint64(len(args.Val)))
	enc[0] = OP_PUT
	enc = marshal.WriteInt(enc, uint64(len(args.Key)))
	enc = marshal.WriteBytes(enc, []byte(args.Key))
	enc = marshal.WriteBytes(enc, []byte(args.Val))
	return enc
}

func DecodePutArgs(raw_args []byte) *PutArgs {
	var enc = raw_args
	args := new(PutArgs)

	var l uint64
	l, enc = marshal.ReadInt(enc)
	args.Key = string(enc[:l])
	args.Val = string(enc[l:])

	return args
}

type getArgs = string

func EncodeGetArgs(args getArgs) []byte {
	var enc = make([]byte, 1, 1+uint64(len(args)))
	enc[0] = OP_GET
	enc = marshal.WriteBytes(enc, []byte(args))
	return enc
}

func decodeGetArgs(raw_args []byte) getArgs {
	return string(raw_args)
}

// end of marshalling
func (s *KVState) put(args *PutArgs) []byte {
	s.kvs[string(args.Key)] = args.Val
	return make([]byte, 0)
}

func (s *KVState) get(args getArgs) []byte {
	return []byte(s.kvs[string(args)])
}

func (s *KVState) apply(args []byte, vnum uint64) []byte {
	if args[0] == OP_PUT {
		args := DecodePutArgs(args[1:])
		s.vnums[string(args.Key)] = vnum
		return s.put(args)
	} else if args[0] == OP_GET {
		key := decodeGetArgs(args[1:])
		s.vnums[string(key)] = vnum
		return s.get(key)
	} else {
		panic("unexpected op type")
	}
}

func (s *KVState) applyReadonly(args []byte) (uint64, []byte) {
	if args[0] != OP_GET {
		panic("expected a GET as readonly-operation")
	}
	key := decodeGetArgs(args[1:])
	reply := s.get(decodeGetArgs(args[1:]))
	vnum, ok := s.vnums[string(key)]
	if ok {
		return vnum, reply
	} else {
		return s.minVnum, reply
	}
}

func (s *KVState) getState() []byte {
	return map_string_marshal.EncodeStringMap(s.kvs)
}

func (s *KVState) setState(snap []byte, nextIndex uint64) {
	s.minVnum = nextIndex
	s.vnums = make(map[string]uint64)

	if len(snap) == 0 {
		s.kvs = make(map[string]string, 0)
	} else {
		s.kvs = map_string_marshal.DecodeStringMap(snap)
	}
}

// func MakeKVStateMachine() *simplelog.InMemoryStateMachine {
// 	s := new(KVState)
// 	s.kvs = make(map[string][]byte, 0)
// 	s.vnums = make(map[string]uint64)
//
// 	return &simplelog.InMemoryStateMachine{
// 		ApplyVolatile: s.apply,
// 		ApplyReadonly: s.applyReadonly,
// 		GetState:      s.getState,
// 		SetState:      s.setState,
// 	}
// }

func makeVersionedStateMachine() *esm.VersionedStateMachine {
	s := new(KVState)
	s.kvs = make(map[string]string, 0)
	s.vnums = make(map[string]uint64)

	return &esm.VersionedStateMachine{
		ApplyVolatile: s.apply,
		ApplyReadonly: s.applyReadonly,
		GetState:      func() []byte { return s.getState() },
		SetState:      s.setState,
	}
}

func Start(fname string, host grove_ffi.Address, confHost grove_ffi.Address) {
	simplelog.MakePbServer(esm.MakeExactlyOnceStateMachine(makeVersionedStateMachine()), fname, confHost).Serve(host)
}
