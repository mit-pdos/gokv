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

	durableNextIndex uint64
	// XXX: this might wake up all the threads at the same time
	durableNextIndex_cond *sync.Cond

	// for read-only operations (primary only)
	committedNextIndex        uint64
	nextRoIndex               uint64
	roOpsToPropose_cond       *sync.Cond
	committedNextRoIndex      uint64
	committedNextRoIndex_cond *sync.Cond
}

func (s *Server) RoApplyAsBackup(args *RoApplyAsBackupArgs) e.Error {
	s.mu.Lock()
	for {
		if args.nextIndex > s.durableNextIndex &&
			s.epoch == args.epoch &&
			s.sealed == false {
			s.durableNextIndex_cond.Wait()
			// FIXME: is one condvar for everything good enough?
		} else {
			break
		}
	}

	if s.epoch != args.epoch {
		s.mu.Unlock()
		return e.Stale
	}
	if s.sealed {
		s.mu.Unlock()
		return e.Sealed
	}
	if args.nextIndex > s.durableNextIndex {
		machine.Assert(false) // shouldn't happen
	}

	s.mu.Unlock()
	return e.None
}

// This commits read-only operations. This is only necessary if RW operations
// are not being applied. If RW ops are being applied, they will implicitly
// commit RO ops.
func (s *Server) applyRoThread(epoch uint64) {
	s.mu.Lock()
	// log.Printf("Started applyRo thread")
	for {
		for {
			if s.epoch != epoch ||
				(s.nextRoIndex != s.committedNextRoIndex &&
					s.nextIndex == s.durableNextIndex) {
				break
			} else {
				// XXX: when increasing durableNextIndex, trigger roOpsToPropse_cond
				s.roOpsToPropose_cond.Wait()
			}
		}
		// log.Printf("About to propose RO ops in applyRo thread")

		// If the server is no longer in the epoch in which this thread is running,
		// then stop this thread.
		if s.epoch != epoch {
			s.mu.Unlock()
			break
		}

		// Else, we can prove that nextRoIndex < committedNextRoIndex, not
		// that it really matters for the following code.
		// Also, nextIndex == durableNextIndex
		nextIndex := s.nextIndex
		nextRoIndex := s.nextRoIndex

		clerks := s.clerks[machine.RandomUint64()%uint64(len(s.clerks))]
		s.mu.Unlock()

		// make a round of RoApplyAsBackup RPCs
		wg := new(sync.WaitGroup)
		args := &RoApplyAsBackupArgs{epoch: epoch, nextIndex: nextIndex}
		errs := make([]e.Error, len(clerks))
		for i, clerk := range clerks {
			i := i
			clerk := clerk
			wg.Add(1)
			go func() {
				for {
					err := clerk.RoApplyAsBackup(args)
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

		// log.Printf("Made RPCs")
		wg.Wait()
		// log.Printf("Done with RPCs")
		// analyze errors
		var err = e.None
		var i = uint64(0)
		for i < uint64(len(errs)) {
			err2 := errs[i]
			if err2 != e.None {
				err = err2
			}
			i += 1
		}
		if err != e.None {
			s.mu.Lock()
			// FIXME: if still in the same epoch and unsealed, then seal
			log.Printf("applyRoThread() exited because of non-retryable RPC error")
			machine.Assume(false)
			s.mu.Unlock()
			break
		}

		s.mu.Lock()
		// If s.nextIndex != nextIndex, then there's a new RW operation in the
		// middle of being proposed/replicated. Whenever that RW operation is
		// committed, this RO op will also be committed, so let the ApplyRo
		// invocation wait for that.
		if s.epoch == epoch && s.nextIndex == nextIndex {
			// log.Printf("New RO ops committed by applyRoThread()")
			s.committedNextRoIndex = nextRoIndex
			s.committedNextRoIndex_cond.Broadcast() // there could be many ApplyRo threads waiting
		}
	}
}

func (s *Server) ApplyRo(op Op) *ApplyReply {
	// primary applying a read-only op
	reply := new(ApplyReply)
	reply.Reply = nil
	// reply.Err = e.ENone
	// return reply
	s.mu.Lock()
	// begin := machine.TimeNow()
	if !s.isPrimary {
		// log.Println("Got read-only request while not being primary")
		s.mu.Unlock()
		reply.Err = e.Stale
		return reply
	}
	if s.sealed {
		s.mu.Unlock()
		reply.Err = e.Stale
		return reply
	}

	// Apply RO op, even though the nextIndex at which it's being applied may
	// not be durable yet.
	reply.Reply = s.sm.ApplyReadonly(op)
	s.nextRoIndex = std.SumAssumeNoOverflow(s.nextRoIndex, 1)
	nextRoIndex := s.nextRoIndex
	nextIndex := s.nextIndex
	epoch := s.epoch
	if s.nextIndex == s.durableNextIndex {
		// Only signal if nextIndex == durableNextIndex; otherwise, let the
		// thread `waitFn()`ing for durability wake up applyRoThread.
		s.roOpsToPropose_cond.Signal()
	}

	for {
		if epoch != s.epoch ||
			s.committedNextIndex >= nextIndex ||
			s.committedNextRoIndex >= nextRoIndex {
			break
		} else {
			s.committedNextRoIndex_cond.Wait()
		}
	}
	s.mu.Unlock()

	if epoch != s.epoch {
		reply.Err = e.Stale
		return reply
	}

	// FIXME: the applyRoThread should quit when getting a non-retryable error
	// (e.g. backup sealed or in new epoch). Account for this case here.

	reply.Err = e.None
	return reply
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

	// now that nextIndex has increased, reset nextRoIndex and committedNextRoIndex
	s.nextRoIndex = 0
	s.committedNextRoIndex = 0
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

	if err == e.None {
		s.mu.Lock()
		sepoch := s.epoch
		if sepoch == epoch && nextIndex > s.committedNextIndex {
			s.committedNextIndex = nextIndex
			s.committedNextRoIndex_cond.Broadcast() // now that committedNextIndex
			// has increased, the outstanding RO ops are implicitly committed.
		}
		s.mu.Unlock()
	}

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

	// FIXME: if we get an index that's smaller than nextIndex, we should just
	// wait for nextIndex to be made durable. That requires saving the waitFn in
	// the server state and making sure that we can call waitFn more than once
	// when its postcondition is persistent.
	// OR: make use of durableNextIndex, which is there for read-only
	// optimization.

	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return e.Stale
	}

	// related to above: Because of the above waiting for args.index to be at most
	// s.nextIndex, if args.index != s.nextIndex, then actually args.index <
	// s.nextIndex and the op has already been accepted in memory, and is being
	// made durable right now.
	//
	// this operation has already been applied, nothing to do.
	if args.index != s.nextIndex {
		s.mu.Unlock()
		return e.OutOfOrder
	}

	// apply it locally
	_, waitFn := s.sm.StartApply(args.op)
	s.nextIndex += 1
	opNextIndex := s.nextIndex

	cond, ok := s.opAppliedConds[s.nextIndex]
	if ok {
		cond.Signal()
		delete(s.opAppliedConds, s.nextIndex)
	}

	s.mu.Unlock()
	waitFn()
	s.mu.Lock()
	if args.epoch == s.epoch && opNextIndex > s.durableNextIndex {
		s.durableNextIndex = opNextIndex
		s.durableNextIndex_cond.Broadcast()

		// after durability, there are some waiting RO ops that might be ok to send to backups
		if s.durableNextIndex == s.nextIndex && s.nextRoIndex != s.committedNextRoIndex {
			s.roOpsToPropose_cond.Signal()
		}
	}
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
		s.nextIndex = args.NextIndex
		s.durableNextIndex = s.nextIndex
		s.sm.SetStateAndUnseal(args.State, args.NextIndex, args.Epoch)

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
		log.Printf("Stale BecomePrimary request (in %d, got %d)", s.epoch, args.Epoch)
		s.mu.Unlock()
		return e.Stale
	}
	log.Println("Became Primary")
	s.isPrimary = true

	// XXX: should probably not bother doing this if we are already the primary
	// in this epoch
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

	// Initialize read-only optimization state
	s.nextRoIndex = 0
	s.committedNextRoIndex = 0
	s.committedNextIndex = 0

	s.mu.Unlock()
	epoch := args.Epoch
	go func() {
		s.applyRoThread(epoch)
	}()

	return e.None
}

func MakeServer(sm *StateMachine, nextIndex uint64, epoch uint64, sealed bool) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.epoch = epoch
	s.sealed = sealed
	s.sm = sm
	s.nextIndex = nextIndex
	s.durableNextIndex = nextIndex
	s.isPrimary = false
	s.opAppliedConds = make(map[uint64]*sync.Cond)
	s.durableNextIndex_cond = sync.NewCond(s.mu)

	s.roOpsToPropose_cond = sync.NewCond(s.mu)
	s.committedNextRoIndex_cond = sync.NewCond(s.mu)
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

	handlers[RPC_ROAPPLYASBACKUP] = func(args []byte, reply *[]byte) {
		*reply = e.EncodeError(s.RoApplyAsBackup(DecodeRoApplyAsBackupArgs(args)))
	}

	handlers[RPC_ROPRIMARYAPPLY] = func(args []byte, reply *[]byte) {
		*reply = EncodeApplyReply(s.ApplyRo(args))
	}

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
