package pb

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
)

type Server struct {
	mu        *sync.Mutex
	epoch     uint64
	sealed    bool
	sm        *StateMachine
	nextIndex uint64

	isPrimary bool
	clerks    [][]*Clerk
	// clerks []*Clerk

	// only on backups
	// opAppliedConds[j] is the condvariable for the op with nextIndex == j.
	opAppliedConds map[uint64]*sync.Cond
}

// called on the primary server to apply a new operation.
func (s *Server) Apply(op Op) *ApplyReply {
	reply := new(ApplyReply)
	reply.Reply = nil
	// reply.Err = e.ENone
	// return reply
	s.mu.Lock()
	// begin := machine.TimeNow()
	if !s.isPrimary {
		// log.Println("Got request while not being primary")
		s.mu.Unlock()
		reply.Err = e.Stale
		return reply
	}
	if s.sealed {
		s.mu.Unlock()
		reply.Err = e.Stale
		return reply
	}

	// apply it locally
	ret, waitForDurable := s.sm.StartApply(op)
	reply.Reply = ret

	nextIndex := s.nextIndex
	s.nextIndex = std.SumAssumeNoOverflow(s.nextIndex, 1)
	epoch := s.epoch
	clerks := s.clerks
	s.mu.Unlock()
	// end := machine.TimeNow()
	// if machine.RandomUint64()%1024 == 0 {
	// log.Printf("replica.mu crit section: %d ns", end-begin)
	// }
	waitForDurable()

	// tell backups to apply it
	wg := new(sync.WaitGroup)
	args := &ApplyAsBackupArgs{
		epoch: epoch,
		index: nextIndex,
		op:    op,
	}

	clerks_inner := clerks[machine.RandomUint64()%uint64(len(clerks))]
	errs := make([]e.Error, len(clerks_inner))
	for i, clerk := range clerks_inner {
		// use a random socket
		clerk := clerk
		i := i
		wg.Add(1)
		go func() {
			// retry if we get OutOfOrder errors
			for {
				err := clerk.ApplyAsBackup(args)
				if err == e.OutOfOrder || err == e.Timeout {
					continue
				} else {
					errs[i] = err
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	var err = e.None
	var i = uint64(0)
	for i < uint64(len(clerks_inner)) {
		err2 := errs[i]
		if err2 != e.None {
			err = err2
		}
		i += 1
	}
	reply.Err = err

	// log.Println("Apply() returned ", err)
	return reply
}

// requires that we've already at least entered this epoch
// returns true iff stale
func (s *Server) isEpochStale(epoch uint64) bool {
	return s.epoch != epoch
}

// called on backup servers to apply an operation so it is replicated and
// can be considered committed by primary.
func (s *Server) ApplyAsBackup(args *ApplyAsBackupArgs) e.Error {
	s.mu.Lock()

	// operation sequencing
	for args.index > s.nextIndex && s.epoch == args.epoch && !s.sealed {
		cond, ok := s.opAppliedConds[args.index]
		if !ok {
			cond := sync.NewCond(s.mu)
			s.opAppliedConds[args.index] = cond
		} else {
			cond.Wait()
		}
	}
	// By this point, if the server is unsealed and in the right epoch, then
	// args.index <= s.nextIndex.
	if s.sealed {
		s.mu.Unlock()
		return e.Stale
	}
	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return e.Stale
	}

	// XXX: Because of the above waiting for args.index to be at most
	// s.nextIndex, if args.index != s.nextIndex, then actually args.index <
	// s.nextIndex and the op has already been accepted. We don't need to prove
	// that, though.
	//
	// this operation has already been applied, nothing to do.
	if args.index != s.nextIndex {
		s.mu.Unlock()
		return e.OutOfOrder
	}

	// apply it locally
	_, waitFn := s.sm.StartApply(args.op)
	s.nextIndex += 1

	cond, ok := s.opAppliedConds[s.nextIndex]
	if ok {
		cond.Signal()
		delete(s.opAppliedConds, s.nextIndex)
	}

	s.mu.Unlock()
	waitFn()

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
		s.nextIndex = args.NextIndex
		s.sm.SetStateAndUnseal(args.State, args.Epoch, args.NextIndex)

		for _, cond := range s.opAppliedConds {
			cond.Signal()
		}
		s.opAppliedConds = make(map[uint64]*sync.Cond)

		s.mu.Unlock()
		return e.None
	}
}

// XXX: probably should rename to GetStateAndSeal
func (s *Server) GetState(args *GetStateArgs) *GetStateReply {
	s.mu.Lock()
	if args.Epoch < s.epoch {
		s.mu.Unlock()
		return &GetStateReply{Err: e.Stale, State: nil}
	}

	s.sealed = true
	ret := s.sm.GetStateAndSeal()
	nextIndex := s.nextIndex

	for _, cond := range s.opAppliedConds {
		cond.Signal()
	}
	s.opAppliedConds = make(map[uint64]*sync.Cond)

	s.mu.Unlock()

	return &GetStateReply{Err: e.None, State: ret, NextIndex: nextIndex}
}

func (s *Server) BecomePrimary(args *BecomePrimaryArgs) e.Error {
	s.mu.Lock()
	// XXX: technically, this != could be a <, and we'd be ok because
	// BecomePrimary can only be called on args.Epoch if the server already
	// entered epoch args.Epoch
	if args.Epoch != s.epoch {
		log.Println("Stale BecomePrimary request")
		s.mu.Unlock()
		return e.Stale
	}
	log.Println("Became Primary")
	s.isPrimary = true

	// XXX: should probably not bother doing this if we are already the primary
	// in this epoch

	/*
		s.clerks = make([]*Clerk, len(args.Replicas)-1)
		var i = uint64(0)
		for i < uint64(len(s.clerks)) {
			s.clerks[i] = MakeClerk(args.Replicas[i+1])
			i++
		}
	*/

	// TODO: multiple sockets
	numClerks := uint64(32) // XXX: 32 clients per backup; this should probably be a configuration parameter
	s.clerks = make([][]*Clerk, numClerks)
	var j = uint64(0)
	for j < numClerks {
		clerks := make([]*Clerk, len(args.Replicas)-1)
		var i = uint64(0)
		for i < uint64(len(clerks)) {
			clerks[i] = MakeClerk(args.Replicas[i+1])
			i++
		}
		s.clerks[j] = clerks
		j++
	}

	s.mu.Unlock()
	return e.None
}

func MakeServer(sm *StateMachine, nextIndex uint64, epoch uint64, sealed bool) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.epoch = epoch
	s.sealed = sealed
	s.sm = sm
	s.nextIndex = nextIndex
	s.isPrimary = false
	s.opAppliedConds = make(map[uint64]*sync.Cond)
	return s
}

func (s *Server) Serve(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[RPC_APPLYASBACKUP] = func(args []byte, reply *[]byte) {
		*reply = e.EncodeError(s.ApplyAsBackup(DecodeApplyAsBackupArgs(args)))
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
		*reply = EncodeApplyReply(s.Apply(args))
	}

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
