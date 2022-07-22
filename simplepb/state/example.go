package state

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/marshal"
	"log"
)

type KVState struct {
	kvs map[uint64][]byte

	filename  string
	epoch     uint64
	nextIndex uint64
}

type Op = []byte

func (s *KVState) loadState(snap_in []byte) {
	log.Println("Loading encoded state: ", len(snap_in))
	var snap = snap_in
	s.kvs = make(map[uint64][]byte, 0)
	numEntries, snap := marshal.ReadInt(snap)
	for i := uint64(0); i < numEntries; i++ {
		var key uint64
		var valLen uint64
		var val []byte
		key, snap = marshal.ReadInt(snap)
		valLen, snap = marshal.ReadInt(snap)
		val = snap[:valLen]
		snap = snap[valLen:]
		s.kvs[key] = val
	}
}

func RecoverKVState(fname string) *KVState {
	s := new(KVState)
	var encState = grove_ffi.Read(fname)
	s.filename = fname
	if len(encState) == 0 {
		s.epoch = 0
		s.nextIndex = 0
		s.kvs = make(map[uint64][]byte)
	} else {
		s.epoch, encState = marshal.ReadInt(encState)
		s.nextIndex, encState = marshal.ReadInt(encState)
		s.loadState(encState)
	}
	return s
}

func (s *KVState) GetState() []byte {
	var enc = make([]byte, 0)
	enc = marshal.WriteInt(enc, uint64(len(s.kvs)))
	for k, v := range s.kvs {
		enc = marshal.WriteInt(enc, k)
		enc = marshal.WriteInt(enc, uint64(len(v)))
		enc = marshal.WriteBytes(enc, v)
	}
	log.Println("Size of encoded state", len(enc))
	return enc
}

func (s *KVState) MakeDurable() {
	state := s.GetState()
	var enc = make([]byte, 0, 16+len(state))

	enc = marshal.WriteInt(enc, s.epoch)
	enc = marshal.WriteInt(enc, s.nextIndex)
	enc = marshal.WriteBytes(enc, state)

	grove_ffi.Write(s.filename, enc)
}

func (s *KVState) Apply(op Op) []byte {
	// the only op is FetchAndAppend(key, val)
	key, appendVal := marshal.ReadInt(op)
	ret := s.kvs[key]
	s.nextIndex += 1
	s.kvs[key] = append(s.kvs[key], appendVal...)

	s.MakeDurable()
	return ret
}

func (s *KVState) SetState(snap_in []byte) {
	s.loadState(snap_in)
	s.MakeDurable()
}

func (s *KVState) EnterEpoch(epoch uint64) {
	s.epoch = epoch
	s.MakeDurable()
}

func MakeKVStateMachine(initState *KVState) *pb.StateMachine {
	return &pb.StateMachine{
		Apply:      initState.Apply,
		SetState:   initState.SetState,
		GetState:   initState.GetState,
		EnterEpoch: initState.EnterEpoch,
	}
}

type KVServer struct {
	r *pb.Server
}

func MakeServer(fname string) *KVServer {
	s := new(KVServer)
	var epoch uint64
	var nextIndex uint64
	state := RecoverKVState(fname)

	s.r = pb.MakeServer(MakeKVStateMachine(state), nextIndex, epoch)
	return s
}

func (s *KVServer) Serve(me grove_ffi.Address) {
	s.r.Serve(me)
}
