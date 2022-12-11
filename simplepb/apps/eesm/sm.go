package eesm

import (
	"github.com/mit-pdos/gokv/map_marshal"
	"github.com/mit-pdos/gokv/simplepb/simplelog"
	"github.com/tchajed/marshal"
)

type EEStateMachine struct {
	lastSeq   map[uint64]uint64
	lastReply map[uint64][]byte
	nextCID   uint64
	sm        *simplelog.InMemoryStateMachine
}

func (s *EEStateMachine) applyVolatile(op []byte) []byte {
	// op[0] is 1 for a GetFreshCID request, and 0 for a normal client op.
	if op[0] == 1 {
		// get fresh cid
		s.nextCID += 1
		var ret = make([]byte, 0, 8)
		ret = marshal.WriteInt(ret, s.nextCID)
		return ret
	} else if op[0] == 0 {
		n := len(op)
		return s.sm.ApplyVolatile(op[1:n])
	}
	panic("unexpected ee op type")
}

func (s *EEStateMachine) getState() []byte {
	appState := s.sm.GetState()
	var enc = make([]byte, 0, uint64(8)+uint64(8)*uint64(len(s.lastSeq))+uint64(len(appState)))

	enc = marshal.WriteInt(enc, s.nextCID)
	enc = map_marshal.EncodeMapU64ToU64(s.lastSeq)
	enc = map_marshal.EncodeMapU64ToBytes(s.lastReply)
	enc = marshal.WriteBytes(enc, appState)

	return enc
}

func (s *EEStateMachine) setState(state []byte) {
	var enc = state
	s.nextCID, enc = marshal.ReadInt(enc)
	s.lastSeq, enc = map_marshal.DecodeMapU64ToU64(enc)
	s.lastReply, enc = map_marshal.DecodeMapU64ToBytes(enc)
	s.sm.SetState(enc)
}

func MakeEEKVStateMachine(sm *simplelog.InMemoryStateMachine) *simplelog.InMemoryStateMachine {
	s := new(EEStateMachine)

	return &simplelog.InMemoryStateMachine{
		ApplyVolatile: s.applyVolatile,
		GetState:      s.getState,
		SetState:      s.setState,
	}
}

func MakeRequest(req []byte) []byte {
	var enc = make([]byte, 1+len(req))
	enc = marshal.WriteBytes(enc, make([]byte, 1))
	enc = marshal.WriteBytes(enc, req)
	return enc
}

func GetCIDRequest() []byte {
	v := make([]byte, 1)
	v[0] = 1
	return v
}
