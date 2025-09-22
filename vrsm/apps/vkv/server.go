package vkv

// Replicated and durable KV server

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/map_string_marshal"
	"github.com/mit-pdos/gokv/vrsm/apps/exactlyonce"
	"github.com/mit-pdos/gokv/vrsm/apps/vkv/condputargs_gk"
	"github.com/mit-pdos/gokv/vrsm/apps/vkv/getargs_gk"
	"github.com/mit-pdos/gokv/vrsm/apps/vkv/putargs_gk"
	"github.com/mit-pdos/gokv/vrsm/storage"
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
	OP_PUT      = byte(0)
	OP_GET      = byte(1)
	OP_COND_PUT = byte(2)
)

func (s *KVState) put(args *putargs_gk.S) []byte {
	s.kvs[args.Key] = args.Val
	return make([]byte, 0)
}

func (s *KVState) get(args getargs_gk.S) []byte {
	return []byte(s.kvs[args.Get])
}

func (s *KVState) apply(args []byte, vnum uint64) []byte {
	if args[0] == OP_PUT {
		args, _ := putargs_gk.Unmarshal(args)
		s.vnums[string(args.Key)] = vnum
		return s.put(&args)
	} else if args[0] == OP_GET {
		key, _ := getargs_gk.Unmarshal(args)
		s.vnums[key.Get] = vnum
		return s.get(key)
	} else if args[0] == OP_COND_PUT {
		args, _ := condputargs_gk.Unmarshal(args)
		if s.kvs[args.Key] == args.Expect {
			s.vnums[string(args.Key)] = vnum
			s.kvs[args.Key] = args.Val
			return []byte("ok")
		}
		return []byte("")
	} else {
		panic("unexpected op type")
	}
}

func (s *KVState) applyReadonly(args []byte) (uint64, []byte) {
	if args[0] != OP_GET {
		panic("expected a GET as readonly-operation")
	}
	key, _ := getargs_gk.Unmarshal(args)
	reply := s.get(key)
	vnum, ok := s.vnums[key.Get]
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
	s.kvs = map_string_marshal.DecodeStringMap(snap)
}

// func MakeKVStateMachine() *storage.InMemoryStateMachine {
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

func makeVersionedStateMachine() *exactlyonce.VersionedStateMachine {
	s := new(KVState)
	s.kvs = make(map[string]string, 0)
	s.vnums = make(map[string]uint64)

	return &exactlyonce.VersionedStateMachine{
		ApplyVolatile: s.apply,
		ApplyReadonly: s.applyReadonly,
		GetState:      func() []byte { return s.getState() },
		SetState:      s.setState,
	}
}

func Start(fname string, host grove_ffi.Address, confHosts []grove_ffi.Address) {
	storage.MakePbServer(exactlyonce.MakeExactlyOnceStateMachine(makeVersionedStateMachine()), fname, confHosts).Serve(host)
}
