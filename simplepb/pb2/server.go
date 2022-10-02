package pb2

/*
import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
)

type StateMachine struct {
	Mu *sync.Mutex
	// these are read-only
	GetEpoch     func() uint64
	GetSealed    func() bool
	GetNextIndex func() uint64

	// apply a new op and get the return value for it
	Apply func(op []byte) []byte
}

type Server struct {
	sm *StateMachine

	isPrimary bool
	// clerks    []*Clerk
}

// called on the primary server to apply a new operation.
func (s *Server) Apply(op []byte) *ApplyReply {
	reply := new(ApplyReply)
	reply.Reply = nil
	s.sm.Mu.Lock()
	if !s.isPrimary {
		log.Println("Got request while not being primary")
		s.sm.Mu.Unlock()
		reply.Err = e.Stale
		return reply
	}
	if s.sm.GetSealed() {
		s.sm.Mu.Unlock()
		reply.Err = e.Stale
		return reply
	}

	// apply it locally
	nextIndex := std.SumAssumeNoOverflow(s.sm.GetNextIndex(), 1)
	epoch := s.sm.GetEpoch()
	reply.Reply = s.sm.Apply(op)

	clerks := s.clerks
	s.sm.Mu.Unlock()

	// FIXME: helper function to do fork/join/reduce
	// tell backups to apply it
	wg := new(sync.WaitGroup)
	errs := make([]e.Error, len(clerks))
	args := &ApplyAsBackupArgs{
		epoch: epoch,
		index: nextIndex,
		op:    op,
	}
	for i, clerk := range clerks {
		clerk := clerk
		i := i
		wg.Add(1)
		go func() {
			errs[i] = clerk.ApplyAsBackup(args)
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
	reply.Err = err

	log.Println("Apply() returned ", err)
	return reply
}

// requires that we've already at least entered this epoch
// returns true iff stale
func (s *Server) IsEpochStale(epoch uint64) bool {
	return s.sm.GetEpoch() > epoch
}

// called on backup servers to apply an operation so it is replicated and
// can be considered committed by primary.
func (s *Server) ApplyAsBackup(args *ApplyAsBackupArgs) e.Error {
	s.sm.Mu.Lock()
	if s.IsEpochStale(args.epoch) {
		s.sm.Mu.Unlock()
		return e.Stale
	}
	if s.sm.GetSealed() {
		s.sm.Mu.Unlock()
		return e.Stale
	}

	if args.index != s.sm.GetNextIndex() {
		s.sm.Mu.Unlock()
		return e.OutOfOrder
	}

	// apply it locally
	s.sm.Apply(args.op)

	s.sm.Mu.Unlock()
	return e.None
}

// func (s *BaseServer) SetState(args *SetStateArgs) e.Error {
// s.mu.Lock()
// if s.epoch > args.Epoch {
// s.mu.Unlock()
// return e.Stale
// } else if s.epoch == args.Epoch {
// s.mu.Unlock()
// return e.None
// } else {
// s.isPrimary = false
// s.epoch = args.Epoch
// s.sealed = false
// s.nextIndex = args.NextIndex
// s.sm.SetStateAndUnseal(args.State, args.Epoch, args.NextIndex)
//
// s.mu.Unlock()
// return e.None
// }
// }
//
// // XXX: probably should rename to GetStateAndSeal
// func (s *BaseServer) GetState(args *GetStateArgs) *GetStateReply {
// s.mu.Lock()
// if s.isEpochStale(args.Epoch) {
// s.mu.Unlock()
// return &GetStateReply{Err: e.Stale, State: nil}
// }
//
// s.sealed = true
// ret := s.sm.GetStateAndSeal()
// nextIndex := s.nextIndex
// s.mu.Unlock()
//
// return &GetStateReply{Err: e.None, State: ret, NextIndex: nextIndex}
// }
func (s *Server) BecomePrimary(args *BecomePrimaryArgs) e.Error {
	s.sm.Mu.Lock()
	if s.IsEpochStale(args.Epoch) {
		log.Println("Stale BecomePrimary request")
		s.sm.Mu.Unlock()
		return e.Stale
	}
	log.Println("Became Primary")
	s.isPrimary = true

	// XXX: should probably not bother doing this if we are already the primary
	// in this epoch
	s.clerks = make([]*Clerk, len(args.Replicas)-1)
	var i = uint64(0)
	for i < uint64(len(s.clerks)) {
		s.clerks[i] = MakeClerk(args.Replicas[i+1])
		i++
	}

	s.sm.Mu.Unlock()
	return e.None
}

func MakeServer(sm *StateMachine) *Server {
}

func (s *Server) Serve(me grove_ffi.Address) {
}
*/
