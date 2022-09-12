package state

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/marshal"
	"log"
)

type KVState struct {
	kvs map[uint64][]byte

	epoch     uint64
	nextIndex uint64
	sealed    bool

	filename string
}

type Op = []byte

// helper for unmarshalling kvs
func decodeKvs(snap_in []byte) map[uint64][]byte {
	log.Println("Decoding encoded state of length: ", len(snap_in))
	var snap = snap_in
	kvs := make(map[uint64][]byte, 0)
	numEntries, snap := marshal.ReadInt(snap)
	for i := uint64(0); i < numEntries; i++ {
		var key uint64
		var valLen uint64
		var val []byte
		key, snap = marshal.ReadInt(snap)
		valLen, snap = marshal.ReadInt(snap)

		// XXX: this will keep the whole original encoded slice around in
		// memory. We probably don't want that.
		val = snap[:valLen]
		snap = snap[valLen:]
		kvs[key] = val
	}
	return kvs
}

func encodeKvs(kvs map[uint64][]byte) []byte {
	var enc = make([]byte, 0)
	enc = marshal.WriteInt(enc, uint64(len(kvs)))
	for k, v := range kvs {
		enc = marshal.WriteInt(enc, k)
		enc = marshal.WriteInt(enc, uint64(len(v)))
		enc = marshal.WriteBytes(enc, v)
	}
	return enc
}

func RecoverKVState(fname string) *KVState {
	s := new(KVState)
	var encState = grove_ffi.Read(fname)
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
		s.kvs = decodeKvs(encState)
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
	enc = marshal.WriteBytes(enc, encodeKvs(s.kvs))
	log.Println("Size of encoded state", len(enc))
	return enc
}

func (s *KVState) MakeDurable() {
	grove_ffi.Write(s.filename, s.getState())
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

func (s *KVState) SetStateAndUnseal(snap_in []byte, epoch uint64, nextIndex uint64) {
	s.kvs = decodeKvs(snap_in)
	s.epoch = epoch
	s.sealed = false
	s.nextIndex = nextIndex
	s.MakeDurable()
}

func (s *KVState) GetStateAndSeal() []byte {
	ret := encodeKvs(s.kvs)
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
		Apply:             initState.Apply,
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
