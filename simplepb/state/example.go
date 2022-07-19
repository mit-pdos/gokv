package state

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type KVState struct {
	kvs map[uint64][]byte

	filename  string
	epoch     uint64
	nextIndex uint64
}

type Op = []byte

func (s *KVState) Apply(op Op) []byte {
	// the only op is FetchAndAppend(key, val)
	key, appendVal := marshal.ReadInt(op)
	ret := s.kvs[key]
	s.nextIndex += 1
	s.kvs[key] = append(s.kvs[key], appendVal...)

	// make stuff durable
	state := s.GetState()
	var enc = make([]byte, 16+len(state))
	marshal.WriteInt(enc, s.epoch)
	marshal.WriteInt(enc, s.nextIndex)
	grove_ffi.Write(s.filename, state)
	return ret
}

func (s *KVState) GetState() []byte {
	enc := make([]byte, 0)
	enc = marshal.WriteInt(enc, uint64(len(s.kvs)))
	for k, v := range s.kvs {
		enc = marshal.WriteInt(enc, k)
		enc = marshal.WriteInt(enc, uint64(len(v)))
		enc = marshal.WriteBytes(enc, v)
	}
	return enc
}

func (s *KVState) SetState(snap []byte) {
	s.kvs = make(map[uint64][]byte, 0)
	numEntries, snap := marshal.ReadInt(snap)
	for i := uint64(0); i < numEntries; i++ {
		var key uint64
		var valLen uint64
		var val []byte
		key, snap = marshal.ReadInt(snap)
		valLen, snap = marshal.ReadInt(snap)
		val, snap = snap[:valLen], snap[valLen:]
		s.kvs[key] = val
	}
}

func MakeKVStateMachine(initState *KVState) *pb.StateMachine {
	return &pb.StateMachine{
		Apply:    initState.Apply,
		SetState: initState.SetState,
		GetState: initState.GetState,
	}
}

type KVServer struct {
	r *pb.Server
}

func (s *KVServer) FetchAndAppend(op []byte) []byte {
	err, ret := s.r.Apply(op)
	if err == pb.ENone {
		var reply = make([]byte, 0, 8+len(ret))
		reply = marshal.WriteInt(reply, err)
		reply = marshal.WriteBytes(reply, ret)
		return reply
	} else {
		var reply = make([]byte, 0, 8)
		reply = marshal.WriteInt(reply, err)
		return reply
	}
}

func MakeKVServer(fname string) *KVServer {
	s := new(KVServer)
	encState := grove_ffi.Read(fname)
	var epoch uint64
	var nextIndex uint64
	var state *KVState
	if len(encState) == 0 {
		epoch = 0
		nextIndex = 0
	} else {
		epoch, encState = marshal.ReadInt(encState)
		nextIndex, encState = marshal.ReadInt(encState)
		state = new(KVState)
		state.SetState(encState)
	}

	s.r = pb.MakeServer(MakeKVStateMachine(state), nextIndex, epoch)
	return s
}

func (s *KVServer) Serve(pbHost grove_ffi.Address, me grove_ffi.Address) {
	s.r.Serve(me)

	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[RPC_FAA] = func(args []byte, reply *[]byte) {
		*reply = s.FetchAndAppend(args)
	}

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
