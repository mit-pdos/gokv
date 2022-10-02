package simplelog

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/marshal"
)

type InMemoryStateMachine struct {
	ApplyVolatile func([]byte) []byte
	GetState      func() []byte
	SetState      func([]byte)
}

func appendOp(fname string, op []byte) {
	var enc = make([]byte, 0, 8+uint64(len(op)))
	// write byte indicating that this is a op, then length of op, then op itself.
	marshal.WriteInt(enc, uint64(len(op)))
	marshal.WriteBytes(enc, op)

	grove_ffi.AtomicAppend(fname, enc)
}

const MAX_LOG_SIZE = uint64(64 * 1024 * 1024)

// File format:
// [N]u8: snapshot
// u64:   epoch
// u64:   nextIndex
// [*]op: ops in the format (op length ++ op)
// ?u8:    sealed; this is only present if the state is sealed in this epoch
type StateMachine struct {
	fname string

	logsize   uint64
	sealed    bool
	epoch     uint64
	nextIndex uint64
	smMem     *InMemoryStateMachine
}

// FIXME: better name; this isn't the same as "MakeDurable"
func (s *StateMachine) makeDurableWithSnap(snap []byte) {
	// TODO: we're copying the entire snapshot in memory just to insert the
	// length before it. Shouldn't do this.
	var enc = make([]byte, 0, 8+len(snap)+8+8)
	marshal.WriteInt(enc, uint64(len(snap)))
	marshal.WriteBytes(enc, snap)
	marshal.WriteInt(enc, s.epoch)
	marshal.WriteInt(enc, s.nextIndex)

	if s.sealed {
		// XXX: maybe we should have a "WriteByte" function?
		marshal.WriteBytes(enc, make([]byte, 1))
	}

	grove_ffi.Write(s.fname, enc)
}

// XXX: this is not safe to run concurrently with apply()
func (s *StateMachine) truncateAndMakeDurable() {
	snap := s.smMem.GetState()
	s.makeDurableWithSnap(snap)
}

func (s *StateMachine) apply(op []byte) []byte {
	ret := s.smMem.ApplyVolatile(op) // apply op in-memory
	s.nextIndex += 1

	s.logsize += uint64(len(op))
	if s.logsize > MAX_LOG_SIZE {
		s.logsize = 0
		s.truncateAndMakeDurable()
	} else {
		appendOp(s.fname, op) // make the op durable
	}
	return ret
}

func (s *StateMachine) setStateAndUnseal(snap []byte, nextIndex uint64, epoch uint64) {
	s.epoch = epoch
	s.nextIndex = nextIndex
	s.sealed = false
	s.smMem.SetState(snap)
	s.makeDurableWithSnap(snap)
}

func (s *StateMachine) getStateAndSeal() []byte {
	if !s.sealed {
		// seal the file by writing a byte at the end
		s.sealed = true
		grove_ffi.AtomicAppend(s.fname, make([]byte, 1))
	}
	// XXX: it might be faster to read the file from disk.
	snap := s.smMem.GetState()
	return snap
}

func recoverStateMachine(smMem *InMemoryStateMachine, fname string) *StateMachine {
	s := &StateMachine{
		fname: fname,
		smMem: smMem,
	}

	// load from file
	var enc = grove_ffi.Read(s.fname)

	// load snapshot
	var snapLen uint64
	var snap []byte
	snapLen, enc = marshal.ReadInt(enc)
	snap = enc[:snapLen]
	enc = enc[snapLen:]
	s.smMem.SetState(snap)

	// load protocol state
	s.epoch, enc = marshal.ReadInt(enc)
	s.nextIndex, enc = marshal.ReadInt(enc)

	// apply ops to bring in-memory state up to date
	for {
		if len(enc) > 1 {
			var opLen uint64
			opLen, enc = marshal.ReadInt(enc)
			op := enc[:opLen]
			enc = enc[opLen:]
			s.smMem.ApplyVolatile(op)
		} else {
			break
		}
	}
	if len(enc) > 0 {
		s.sealed = true
	}

	return s
}

// XXX: putting this here because MakeServer takes nextIndex, epoch, and sealed
// as input, and the user of simplelog won't have access to the private fields
// index, epoch, etc.
//
// Maybe we should make those be a part of pb.StateMachine
func MakePbServer(smMem *InMemoryStateMachine, fname string) *pb.Server {
	s := recoverStateMachine(smMem, fname)
	sm := &pb.StateMachine{
		Apply:             s.apply,
		SetStateAndUnseal: s.setStateAndUnseal,
		GetStateAndSeal:   s.getStateAndSeal,
	}
	return pb.MakeServer(sm, s.nextIndex, s.epoch, s.sealed)
}
