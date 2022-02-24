package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
	"sync"
)

type VersionedValue struct {
	ver uint64
	val []byte
}

type ConfServer struct {
	mu  *sync.Mutex
	kvs map[uint64]VersionedValue
}

const CONF_PUT = uint64(1)
const CONF_GET = uint64(1)

type PutArgs struct {
	key     uint64
	prevVer uint64
	newVal  []byte
}

// MARSHAL
func EncodePutArgs(args *PutArgs) []byte {
	enc := marshal.NewEnc(8 + 8 + uint64(len(args.newVal)))
	enc.PutInt(args.key)
	enc.PutInt(args.prevVer)
	enc.PutBytes(args.newVal)
	return enc.Finish()
}

// MARSHAL
func DecodePutArgs(data []byte) *PutArgs {
	dec := marshal.NewDec(data)
	args := new(PutArgs)
	args.key = dec.GetInt()
	args.prevVer = dec.GetInt()
	args.newVal = dec.GetBytes(uint64(len(data)) - 16) // FIXME: annoying marshal interface
	return args
}

func EncodeVersionedValue(v *VersionedValue) []byte {
	enc := marshal.NewEnc(8 + uint64(len(v.val)))
	enc.PutInt(v.ver)
	enc.PutBytes(v.val)
	return enc.Finish()
}

func DecodeVersionedValue(data []byte) *VersionedValue {
	dec := marshal.NewDec(data)
	v := new(VersionedValue)
	v.ver = dec.GetInt()
	v.val = dec.GetBytes(uint64(len(data)) - 8)
	return v
}

func (s *ConfServer) PutRPC(args *PutArgs) bool {
	s.mu.Lock()
	_, ok := s.kvs[args.key]
	if ok {
		if s.kvs[args.key].ver == args.prevVer {
			s.kvs[args.key] = VersionedValue{ver: args.prevVer + 1, val: args.newVal}
		}
	} else {
		s.kvs[args.key] = VersionedValue{ver: args.prevVer + 1, val: args.newVal}
	}
	s.mu.Unlock()
	return true
}

type GetReply struct {
	ver uint64
	val []byte
}

func (s *ConfServer) GetRPC(key uint64, v *VersionedValue) {
	s.mu.Lock()
	*v = s.kvs[key]
	s.mu.Unlock()
}

func StartConfServer(me grove_ffi.Address) {
	s := new(ConfServer)
	s.mu = new(sync.Mutex)
	s.kvs = make(map[uint64]VersionedValue)

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[CONF_PUT] = func(args []byte, rep *[]byte) {
		if s.PutRPC(DecodePutArgs(args)) {
			*rep = make([]byte, 1)
		} else {
			*rep = make([]byte, 0)
		}
	}

	handlers[CONF_GET] = func(args []byte, rep *[]byte) {
		v := new(VersionedValue)
		s.GetRPC(machine.UInt64Get(args), v)
	}

	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)
}

type ConfClerk struct {
	cl *rpc.RPCClient
}

func (c *ConfClerk) Put(key, prevVer uint64, newVal []byte) bool {
	raw_reply := new([]byte)
	raw_args := EncodePutArgs(&PutArgs{key: key, prevVer: prevVer, newVal: newVal})
	err := c.cl.Call(CONF_PUT, raw_args, raw_reply, 100 /* ms */)
	if err == 0 { // FIXME: add ENone or some such
		return (uint64(len(*raw_reply)) > 0)
	}
	return false
}

func (c *ConfClerk) Get(key uint64) *VersionedValue {
	raw_reply := new([]byte)
	raw_args := make([]byte, 8)
	machine.UInt64Put(raw_args, key)

	err := c.cl.Call(CONF_GET, raw_args, raw_reply, 100 /* ms */)
	if err == 0 {
		return DecodeVersionedValue(*raw_reply)
	}
	// FIXME: else retry or report error
	machine.Assume(false)
	return nil
}

func MakeConfClerk(confServer grove_ffi.Address) *ConfClerk {
	return &ConfClerk{cl: rpc.MakeRPCClient(confServer)}
}
