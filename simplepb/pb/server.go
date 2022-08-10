package pb

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type StateMachine struct {
	Apply           func(op Op) []byte
	SetState        func(snap []byte, nextIndex uint64, epoch uint64)
	GetStateAndSeal func() []byte
}

type Server struct {
	mu        *sync.Mutex
	epoch     uint64
	sealed    bool
	sm        *StateMachine
	nextIndex uint64

	isPrimary bool
	clerks    []*Clerk
}

// called on the primary server to apply a new operation.
func (s *Server) Apply(op Op) (e.Error, []byte) {
	s.mu.Lock()
	if !s.isPrimary {
		log.Println("Got request while not being primary")
		s.mu.Unlock()
		return e.Stale, nil
	}
	if s.sealed {
		s.mu.Unlock()
		return e.Stale, nil // TODO: make e.Sealed
	}

	// apply it locally
	ret := s.sm.Apply(op)

	nextIndex := s.nextIndex
	s.nextIndex = std.SumAssumeNoOverflow(s.nextIndex, 1)
	epoch := s.epoch
	clerks := s.clerks
	s.mu.Unlock()

	// tell backups to apply it
	wg := new(sync.WaitGroup)
	errs := make([]e.Error, len(clerks))
	args := &ApplyArgs{
		epoch: epoch,
		index: nextIndex,
		op:    op,
	}
	for i, clerk := range clerks {
		clerk := clerk
		i := i
		wg.Add(1)
		go func() {
			errs[i] = clerk.Apply(args)
			wg.Done()
		}()
	}
	wg.Wait()
	var err = e.None
	var i = uint64(0)
	for i < uint64(len(clerks)) {
		err2 := errs[i]
		if err2 != e.None {
			err = err2
		}
		i += 1
	}

	log.Println("Apply() returned ", err)
	return err, ret
}

// requires that we've already at least entered this epoch
// returns true iff stale
func (s *Server) isEpochStale(epoch uint64) bool {
	return s.epoch > epoch
}

// called on backup servers to apply an operation so it is replicated and
// can be considered committed by primary.
func (s *Server) ApplyAsBackup(args *ApplyArgs) e.Error {
	s.mu.Lock()
	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return e.Stale
	}
	if s.sealed {
		s.mu.Unlock()
		return e.Stale
	}

	if args.index != s.nextIndex {
		s.mu.Unlock()
		return e.OutOfOrder
	}

	// apply it locally
	s.sm.Apply(args.op)
	s.nextIndex += 1

	s.mu.Unlock()
	return e.None
}

func (s *Server) SetState(args *SetStateArgs) e.Error {
	s.mu.Lock()
	if s.epoch > args.Epoch {
		s.mu.Unlock()
		return e.Stale
	} else if s.epoch == args.Epoch {
		s.mu.Unlock()
		return e.None
	} else {
		s.isPrimary = false
		s.epoch = args.Epoch
		s.sealed = false
		s.sm.SetState(args.State, args.Epoch, args.NextIndex)

		s.mu.Unlock()
		return e.None
	}
}

// XXX: probably should rename to GetStateAndSeal
func (s *Server) GetState(args *GetStateArgs) *GetStateReply {
	s.mu.Lock()
	if s.isEpochStale(args.Epoch) {
		s.mu.Unlock()
		return &GetStateReply{Err: e.Stale, State: nil}
	}

	s.sealed = true
	ret := s.sm.GetStateAndSeal()
	nextIndex := s.nextIndex
	s.mu.Unlock()

	return &GetStateReply{Err: e.None, State: ret, NextIndex: nextIndex}
}

func (s *Server) BecomePrimary(args *BecomePrimaryArgs) e.Error {
	s.mu.Lock()
	if s.isEpochStale(args.Epoch) {
		log.Println("Stale BecomePrimary request")
		s.mu.Unlock()
		return e.Stale
	}
	log.Println("Became Primary")
	s.isPrimary = true

	s.clerks = make([]*Clerk, len(args.Replicas)-1)
	for i := range s.clerks {
		s.clerks[i] = MakeClerk(args.Replicas[i+1])
	}

	s.mu.Unlock()
	return e.None
}

func MakeServer(sm *StateMachine, nextIndex uint64, epoch uint64) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.epoch = epoch
	s.sm = sm
	s.nextIndex = nextIndex
	s.isPrimary = false
	return s
}

func (s *Server) Serve(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[RPC_APPLY] = func(args []byte, reply *[]byte) {
		*reply = e.EncodeError(s.ApplyAsBackup(DecodeApplyArgs(args)))
	}

	handlers[RPC_SETSTATE] = func(args []byte, reply *[]byte) {
		*reply = e.EncodeError(s.SetState(DecodeSetStateArgs(args)))
	}

	handlers[RPC_GETSTATE] = func(args []byte, reply *[]byte) {
		*reply = EncodeGetStateReply(s.GetState(DecodeGetStateArgs(args)))
	}

	handlers[RPC_BECOMEPRIMARY] = func(args []byte, reply *[]byte) {
		*reply = e.EncodeError(s.BecomePrimary(DecodeBecomePrimaryArgs(args)))
	}

	handlers[RPC_PRIMARYAPPLY] = func(args []byte, reply *[]byte) {
		err, ret := s.Apply(args)
		if err == e.None {
			*reply = make([]byte, 0, 8+len(ret))
			*reply = marshal.WriteInt(*reply, err)
			*reply = marshal.WriteBytes(*reply, ret)
		} else {
			*reply = make([]byte, 0, 8)
			*reply = marshal.WriteInt(*reply, err)
		}
	}

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
