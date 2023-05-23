package pb

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
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

	// This prevents a primary from becoming primary in the same epoch again
	// after a crash. This allows the primary to send RPCs to backups while
	// still waiting for ops to be made locally durable.
	canBecomePrimary bool
	isPrimary        bool
	clerks           [][]*Clerk

	isPrimary_cond *sync.Cond

	// only on backups
	// opAppliedConds[j] is the condvariable for the op with nextIndex == j.
	opAppliedConds map[uint64]*sync.Cond

	// for read-only operations (primary only)
	leaseExpiration         uint64
	leaseValid              bool
	committedNextIndex      uint64
	committedNextIndex_cond *sync.Cond
	confCk                  *config.Clerk
}

// Applies the RO op immediately, but then waits for it to be committed before
// replying to client.
func (s *Server) ApplyRoWaitForCommit(op Op) *ApplyReply {
	reply := new(ApplyReply)
	reply.Reply = nil
	reply.Err = e.None

	// x := machine.RandomUint64()
	// log.Printf("Got ro request %d", x)
	s.mu.Lock()
	if !s.leaseValid {
		s.mu.Unlock()
		log.Printf("Lease invalid")
		reply.Err = e.LeaseExpired
		return reply
	}
	if machine.RandomUint64()%10000 == 0 {
		log.Printf("Server nextIndex=%d commitIndex=%d", s.nextIndex, s.committedNextIndex)
	}

	var lastModifiedIndex uint64
	lastModifiedIndex, reply.Reply = s.sm.ApplyReadonly(op)
	epoch := s.epoch

	_, h := grove_ffi.GetTimeRange()
	if s.leaseExpiration <= h {
		s.mu.Unlock()
		log.Printf("Lease expired because %d < %d", s.leaseExpiration, h)
		reply.Err = e.LeaseExpired
		return reply
	}

	for {
		if s.epoch != epoch {
			reply.Err = e.Stale
			break
		} else if lastModifiedIndex <= s.committedNextIndex {
			reply.Err = e.None
			break
		} else {
			s.committedNextIndex_cond.Wait()
			continue
		}
	}
	s.mu.Unlock()

	// log.Printf("Success ro request %d", x)
	return reply
}

// precondition:
// is_epoch_lb epoch ∗ committed_by epoch log ∗ is_pb_log_lb log
func (s *Server) IncreaseCommitIndex(newCommittedNextIndex uint64) {
	s.mu.Lock()
	if newCommittedNextIndex > s.committedNextIndex && newCommittedNextIndex <= s.nextIndex {
		s.committedNextIndex = newCommittedNextIndex
		s.committedNextIndex_cond.Broadcast() // now that committedNextIndex
		// has increased, the outstanding RO ops are complete.
	}
	s.mu.Unlock()
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

	opIndex := s.nextIndex
	s.nextIndex = std.SumAssumeNoOverflow(s.nextIndex, 1)
	nextIndex := s.nextIndex
	epoch := s.epoch
	clerks := s.clerks

	s.mu.Unlock()
	// end := machine.TimeNow()
	// if machine.RandomUint64()%1024 == 0 {
	// log.Printf("replica.mu crit section: %d ns", end-begin)
	// }

	// tell backups to apply it
	wg := new(sync.WaitGroup)
	args := &ApplyAsBackupArgs{
		epoch: epoch,
		index: opIndex,
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
				// log.Printf("Sending applyasbackup")
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

	// log.Printf("wait durable: %d", nextIndex)
	waitForDurable()
	// log.Printf("done durable: %d", nextIndex)

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
		s.IncreaseCommitIndex(nextIndex)
	} else {
		// stop acting as primary in epoch
		s.mu.Lock()
		if s.epoch == epoch {
			s.isPrimary = false
		}
		s.mu.Unlock()
	}

	// log.Println("Apply() returned ", err)
	return reply
}

func (s *Server) leaseRenewalThread() {
	var latestEpoch uint64
	for {
		leaseErr, leaseExpiration := s.confCk.GetLease(latestEpoch)

		s.mu.Lock()
		if s.epoch == latestEpoch && leaseErr == e.None {
			s.leaseExpiration = leaseExpiration
			s.leaseValid = true
			s.mu.Unlock()
			// log.Printf("Got lease")
			machine.Sleep(uint64(250) * 1_000_000)
		} else if latestEpoch != s.epoch {
			latestEpoch = s.epoch
			s.mu.Unlock()
		} else { // XXX: in this case, got an error but we're still in the same epoch.
			// This happens e.g. when the config service wants to expire the
			// lease to move to a new epoch. We should avoid sending requests
			// too quickly to the config service in that case
			s.mu.Unlock()
			machine.Sleep(uint64(50) * 1_000_000)
		}
	}
}

func (s *Server) sendIncreaseCommitThread() {
	// XXX: this will keep sending IncreaseCommitIndex RPCs.
	// One might think that we could do this only when committedNextIndex
	// increases, but a backup might crash and restart and forget what its
	// committedNextIndex was before. In that case, with no writes going on,
	// either we would need to add a mechanism for backups to ask the primary
	// for the latest committedNextIndex, or else have the primary proactively
	// push it to the backups. We chose option 2 here
	for {
		s.mu.Lock()
		for !s.isPrimary || len(s.clerks[0]) == 0 {
			s.isPrimary_cond.Wait()
		}
		newCommittedNextIndex := s.committedNextIndex
		clerks := s.clerks
		s.mu.Unlock()

		clerks_inner := clerks[machine.RandomUint64()%uint64(len(clerks))] // use a random socket
		wg := new(sync.WaitGroup)
		for _, clerk := range clerks_inner {
			clerk := clerk
			wg.Add(1)
			go func() {
				// retry if we get error, to make sure every backup gets brought up to date
				for {
					err := clerk.IncreaseCommitIndex(newCommittedNextIndex)
					if err == e.None {
						break
					} else {
						continue
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		// XXX: this is so the primary does not flood the backups with RPCs
		// (e.g. when the system should be idle).
		machine.Sleep(5_000_000) // 5 ms
	}
}

// requires that we've already at least entered this epoch
// returns true iff stale
func (s *Server) isEpochStale(epoch uint64) bool {
	return s.epoch != epoch
}

// called on backup servers to apply an operation so it is replicated and
// can be considered committed by primary.
func (s *Server) ApplyAsBackup(args *ApplyAsBackupArgs) e.Error {
	// log.Printf("Received applyasbackup")
	// defer log.Printf("Exiting applyasbackup")
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
		log.Print("Entered new epoch")
		s.isPrimary = false
		s.canBecomePrimary = true
		s.epoch = args.Epoch
		s.leaseValid = false
		s.sealed = false
		s.nextIndex = args.NextIndex
		s.sm.SetStateAndUnseal(args.State, args.NextIndex, args.Epoch)

		for _, cond := range s.opAppliedConds {
			cond.Signal()
		}
		s.committedNextIndex_cond.Broadcast()
		s.opAppliedConds = make(map[uint64]*sync.Cond)

		s.mu.Unlock()
		s.IncreaseCommitIndex(args.CommittedNextIndex)
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
	committedNextIndex := s.committedNextIndex

	for _, cond := range s.opAppliedConds {
		cond.Signal()
	}
	s.opAppliedConds = make(map[uint64]*sync.Cond)
	s.committedNextIndex_cond.Broadcast()
	s.mu.Unlock()

	return &GetStateReply{Err: e.None, State: ret, NextIndex: nextIndex,
		CommittedNextIndex: committedNextIndex}
}

func (s *Server) BecomePrimary(args *BecomePrimaryArgs) e.Error {
	s.mu.Lock()
	// XXX: technically, this != could be a <, and we'd be ok because
	// BecomePrimary can only be called on args.Epoch if the server already
	// entered epoch args.Epoch
	if args.Epoch != s.epoch || !s.canBecomePrimary {
		log.Printf("Wrong epoch in BecomePrimary request (in %d, got %d)", s.epoch, args.Epoch)
		s.mu.Unlock()
		return e.Stale
	}
	log.Println("Became Primary")
	s.isPrimary = true
	s.isPrimary_cond.Signal()
	s.canBecomePrimary = false

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
	s.mu.Unlock()
	return e.None
}

func MakeServer(sm *StateMachine, confHost grove_ffi.Address, nextIndex uint64, epoch uint64, sealed bool) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.epoch = epoch
	s.sealed = sealed
	s.sm = sm
	s.nextIndex = nextIndex
	s.isPrimary = false
	s.canBecomePrimary = false
	s.leaseValid = false
	s.canBecomePrimary = false
	s.opAppliedConds = make(map[uint64]*sync.Cond)
	s.confCk = config.MakeClerk(confHost)
	s.committedNextIndex_cond = sync.NewCond(s.mu)
	s.isPrimary_cond = sync.NewCond(s.mu)

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

	handlers[RPC_ROPRIMARYAPPLY] = func(args []byte, reply *[]byte) {
		*reply = EncodeApplyReply(s.ApplyRoWaitForCommit(args))
	}

	handlers[RPC_INCREASECOMMIT] = func(args []byte, reply *[]byte) {
		s.IncreaseCommitIndex(DecodeIncreaseCommitArgs(args))
	}

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)

	go func() { s.leaseRenewalThread() }()
	go func() { s.sendIncreaseCommitThread() }()
}
