package storage

import (
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/aof"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/vrsm/replica"
	"github.com/tchajed/marshal"
)

type InMemoryStateMachine struct {
	ApplyReadonly func([]byte) (uint64, []byte)
	ApplyVolatile func([]byte) []byte
	GetState      func() []byte
	SetState      func([]byte, uint64)
}

const MAX_LOG_SIZE = uint64(64 * 1024 * 1024 * 1024)

// File format:
// [N]u8: snapshot
// u64:   epoch
// u64:   nextIndex
// [*]op: ops in the format (op length ++ op)
// ?u8:    sealed; this is only present if the state is sealed in this epoch
type StateMachine struct {
	fname string

	// this append-only file
	logFile *aof.AppendOnlyFile

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
	enc = marshal.WriteInt(enc, uint64(len(snap)))
	enc = marshal.WriteBytes(enc, snap)
	enc = marshal.WriteInt(enc, s.epoch)
	enc = marshal.WriteInt(enc, s.nextIndex)

	if s.sealed {
		// XXX: maybe we should have a "WriteByte" function?
		marshal.WriteBytes(enc, make([]byte, 1))
	}

	s.logFile.Close()
	grove_ffi.FileWrite(s.fname, enc)
	s.logFile = aof.CreateAppendOnlyFile(s.fname)
}

// XXX: this is not safe to run concurrently with apply()
// requires that the state machine is not sealed
func (s *StateMachine) truncateAndMakeDurable() {
	snap := s.smMem.GetState()
	s.makeDurableWithSnap(snap)
}

func (s *StateMachine) apply(op []byte) ([]byte, func()) {
	ret := s.smMem.ApplyVolatile(op) // apply op in-memory
	s.nextIndex = std.SumAssumeNoOverflow(s.nextIndex, 1)

	s.logsize += uint64(len(op))

	// if s.logsize > MAX_LOG_SIZE {
	// panic("unsupported when using aof")
	// s.logsize = 0
	// s.truncateAndMakeDurable()
	// } else {
	var opWithLen = make([]byte, 0, 8+uint64(len(op)))
	opWithLen = marshal.WriteInt(opWithLen, uint64(len(op)))
	opWithLen = marshal.WriteBytes(opWithLen, op)
	l := s.logFile.Append(opWithLen)

	// XXX: need to read this outside the goroutine because the logFile
	// might be deleted and a new one take it place.
	f := s.logFile
	waitFn := func() {
		f.WaitAppend(l)
	}
	return ret, waitFn
	// }
}

func (s *StateMachine) applyReadonly(op []byte) (uint64, []byte) {
	return s.smMem.ApplyReadonly(op) // apply op in-memory
}

// TODO: make the nextIndex and epoch argument order consistent with replica.StateMachine
func (s *StateMachine) setStateAndUnseal(snap []byte, nextIndex uint64, epoch uint64) {
	s.epoch = epoch
	s.nextIndex = nextIndex
	s.sealed = false
	s.smMem.SetState(snap, nextIndex)
	s.makeDurableWithSnap(snap)
}

func (s *StateMachine) getStateAndSeal() []byte {
	// if sealed, then _definitely_ have up-to-date resources
	if !s.sealed {
		// seal the file by writing a byte at the end
		s.sealed = true
		l := s.logFile.Append(make([]byte, 1))
		s.logFile.WaitAppend(l)
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
	var enc = grove_ffi.FileRead(s.fname)

	if len(enc) == 0 {
		// this means the file represents an empty snapshot, epoch 0, and nextIndex 0
		// write that in the file to start
		initState := smMem.GetState()

		var initialContents = make([]byte, 0, 8+uint64(len(initState))+8+8)
		initialContents = marshal.WriteInt(initialContents, uint64(len(initState)))
		initialContents = marshal.WriteBytes(initialContents, initState)
		initialContents = marshal.WriteInt(initialContents, 0)
		initialContents = marshal.WriteInt(initialContents, 0)

		grove_ffi.FileWrite(s.fname, initialContents)

		s.logFile = aof.CreateAppendOnlyFile(fname)
		return s
	}

	s.logFile = aof.CreateAppendOnlyFile(fname)

	// load snapshot
	var snapLen uint64
	var snap []byte
	snapLen, enc = marshal.ReadInt(enc)
	snap = enc[0:snapLen]
	n := len(enc) // For `make check`
	enc = enc[snapLen:n]

	s.epoch, enc = marshal.ReadInt(enc)
	s.nextIndex, enc = marshal.ReadInt(enc)

	s.smMem.SetState(snap, s.nextIndex)

	// load protocol state

	// apply ops to bring in-memory state up to date
	for {
		// XXX: this depends on the fact that an `op` takes up at least 2 bytes
		// e.g. because its opLen takes 8 bytes. A single extra byte is
		// considered a "sealed" flag.
		if len(enc) > 1 {
			var opLen uint64
			opLen, enc = marshal.ReadInt(enc)
			op := enc[0:opLen]
			n := len(enc)
			enc = enc[opLen:n]
			s.smMem.ApplyVolatile(op)
			s.nextIndex = std.SumAssumeNoOverflow(s.nextIndex, 1)
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
// Maybe we should make those be a part of replica.StateMachine
func MakePbServer(smMem *InMemoryStateMachine, fname string, confHosts []grove_ffi.Address) *replica.Server {
	s := recoverStateMachine(smMem, fname)
	sm := &replica.StateMachine{
		StartApply: func(op []byte) ([]byte, func()) {
			return s.apply(op)
		},
		ApplyReadonly: func(op []byte) (uint64, []byte) {
			return s.applyReadonly(op)
		},
		SetStateAndUnseal: func(snap []byte, nextIndex uint64, epoch uint64) {
			s.setStateAndUnseal(snap, nextIndex, epoch)
		},
		GetStateAndSeal: func() []byte {
			return s.getStateAndSeal()
		},
	}
	return replica.MakeServer(sm, confHosts, s.nextIndex, s.epoch, s.sealed)
}
