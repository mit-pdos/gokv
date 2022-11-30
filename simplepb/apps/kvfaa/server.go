package kvfaa

import (
	"log"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/map_marshal"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/marshal"
)

type KVState struct {
	kvs map[uint64][]byte

	epoch     uint64
	nextIndex uint64
	sealed    bool

	filename string
}

type Op = []byte

func RecoverKVState(fname string) *KVState {
	s := new(KVState)
	var encState = grove_ffi.FileRead(fname)
	s.filename = fname
	if len(encState) == 0 {
		s.epoch = 0
		s.nextIndex = 0
		s.sealed = false
		s.kvs = make(map[uint64][]byte)
	} else {
		s.epoch, encState = marshal.ReadInt(encState)
		s.nextIndex, encState = marshal.ReadInt(encState)

		var sealedInt uint64
		sealedInt, encState = marshal.ReadInt(encState)
		s.sealed = (sealedInt == 0)
		log.Println("Decoding encoded state of length: ", len(encState))
		s.kvs = map_marshal.DecodeMapU64ToBytes(encState)
	}
	return s
}

func (s *KVState) getState() []byte {
	var enc = make([]byte, 0)
	enc = marshal.WriteInt(enc, s.epoch)
	enc = marshal.WriteInt(enc, s.nextIndex)
	if s.sealed {
		enc = marshal.WriteInt(enc, 1)
	} else {
		enc = marshal.WriteInt(enc, 0)
	}
	enc = marshal.WriteBytes(enc, map_marshal.EncodeMapU64ToBytes(s.kvs))
	log.Println("Size of encoded state", len(enc))
	return enc
}

func (s *KVState) MakeDurable() {
	grove_ffi.FileWrite(s.filename, s.getState())
}

func (s *KVState) Apply(op Op) ([]byte, func()) {
	// the only op is FetchAndAppend(key, val)
	key, appendVal := marshal.ReadInt(op)
	ret := s.kvs[key]
	s.nextIndex = std.SumAssumeNoOverflow(s.nextIndex, 1)
	s.kvs[key] = append(s.kvs[key], appendVal...)

	s.MakeDurable()
	return ret, func() {}
}

func (s *KVState) SetStateAndUnseal(snap_in []byte, epoch uint64, nextIndex uint64) {
	s.kvs = map_marshal.DecodeMapU64ToBytes(snap_in)
	s.epoch = epoch
	s.sealed = false
	s.nextIndex = nextIndex
	s.MakeDurable()
}

func (s *KVState) GetStateAndSeal() []byte {
	ret := map_marshal.EncodeMapU64ToBytes(s.kvs)
	s.sealed = true
	s.MakeDurable()
	return ret
}

func (s *KVState) EnterEpoch(epoch uint64) {
	s.epoch = epoch
	s.MakeDurable()
}

func MakeKVStateMachine(initState *KVState) *pb.StateMachine {
	return &pb.StateMachine{
		StartApply:        initState.Apply,
		SetStateAndUnseal: initState.SetStateAndUnseal,
		GetStateAndSeal:   initState.GetStateAndSeal,
	}
}

type KVServer struct {
	r *pb.Server
}

func MakeServer(fname string) *KVServer {
	s := new(KVServer)
	state := RecoverKVState(fname)

	s.r = pb.MakeServer(MakeKVStateMachine(state), state.nextIndex, state.epoch, state.sealed)
	return s
}

func (s *KVServer) Serve(me grove_ffi.Address) {
	s.r.Serve(me)
}
