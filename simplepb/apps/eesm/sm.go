package eesm

import (
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/map_marshal"
	"github.com/mit-pdos/gokv/simplepb/clerk"
	"github.com/mit-pdos/gokv/simplepb/simplelog"
	"github.com/tchajed/marshal"
)

type EEStateMachine struct {
	lastSeq   map[uint64]uint64
	lastReply map[uint64][]byte
	nextCID   uint64
	sm        *simplelog.InMemoryStateMachine
}

const (
	OPTYPE_RW          = byte(0)
	OPTYPE_GETFRESHCID = byte(1)
	OPTYPE_RO          = byte(2)
)

func (s *EEStateMachine) applyVolatile(op []byte) []byte {
	var ret []byte
	// op[0] is 1 for a GetFreshCID request, 0 for a RW op, and 2 for an RO op.
	if op[0] == OPTYPE_GETFRESHCID {
		// get fresh cid
		ret = make([]byte, 0, 8)
		ret = marshal.WriteInt(ret, s.nextCID)
		s.nextCID += 1
	} else if op[0] == OPTYPE_RW {
		n := len(op)
		enc := op[1:n]
		cid, enc2 := marshal.ReadInt(enc)
		seq, realOp := marshal.ReadInt(enc2)

		if s.lastSeq[cid] >= seq {
			ret = s.lastReply[cid]
		} else {
			ret = s.sm.ApplyVolatile(realOp)
			s.lastReply[cid] = ret
			s.lastSeq[cid] = seq
		}
	} else if op[0] == OPTYPE_RO {
		n := len(op)
		realOp := op[1:n]
		ret = s.sm.ApplyReadonly(realOp)
	} else {
		panic("unexpected ee op type")
	}
	return ret
}

func (s *EEStateMachine) applyReadonly(op []byte) []byte {
	var ret []byte
	// op[0] is 1 for a GetFreshCID request, 0 for a RW op, and 2 for an RO op.
	if op[0] == OPTYPE_GETFRESHCID {
		panic("Got GETFRESHCID as a read-only op")
	} else if op[0] == OPTYPE_RW {
		panic("Got RW as a read-only op")
	} else if op[0] == OPTYPE_RO {
		n := len(op)
		realOp := op[1:n]
		ret = s.sm.ApplyReadonly(realOp)
	} else {
		panic("unexpected ee op type")
	}
	return ret
}

func (s *EEStateMachine) getState() []byte {
	appState := s.sm.GetState()
	// var enc = make([]byte, 0, uint64(8)+uint64(8)*uint64(len(s.lastSeq))+uint64(len(appState)))
	var enc = make([]byte, 0, 0) // XXX: reservation causes potential overflow in proof

	enc = marshal.WriteInt(enc, s.nextCID)
	enc = marshal.WriteBytes(enc, map_marshal.EncodeMapU64ToU64(s.lastSeq))
	enc = marshal.WriteBytes(enc, map_marshal.EncodeMapU64ToBytes(s.lastReply))
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
	s.lastSeq = make(map[uint64]uint64)
	s.lastReply = make(map[uint64][]byte)
	s.nextCID = 0
	s.sm = sm

	return &simplelog.InMemoryStateMachine{
		ApplyVolatile: s.applyVolatile,
		ApplyReadonly: s.applyReadonly,
		GetState:      func() []byte { return s.getState() },
		SetState:      s.setState,
	}
}

type Clerk struct {
	ck  *clerk.Clerk
	cid uint64
	seq uint64
}

func MakeClerk(confHost grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.ck = clerk.Make(confHost)

	v := make([]byte, 1)
	v[0] = OPTYPE_GETFRESHCID
	cidEnc := ck.ck.Apply(v)
	ck.cid, _ = marshal.ReadInt(cidEnc)
	ck.seq = 1
	return ck
}

func (ck *Clerk) ApplyExactlyOnce(req []byte) []byte {
	var enc = make([]byte, 1, 1) // XXX: reservation causes potential overflow in proof
	enc[0] = OPTYPE_RW
	enc = marshal.WriteInt(enc, ck.cid)
	enc = marshal.WriteInt(enc, ck.seq)
	enc = marshal.WriteBytes(enc, req)
	ck.seq = std.SumAssumeNoOverflow(ck.seq, 1)

	return ck.ck.Apply(enc)
}

func (ck *Clerk) ApplyReadonly(req []byte) []byte {
	var enc = make([]byte, 1, 1) // XXX: reservation causes potential overflow in proof
	enc[0] = OPTYPE_RO
	enc = marshal.WriteBytes(enc, req)
	return ck.ck.ApplyRo(enc)
}
