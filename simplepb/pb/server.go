package pb

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
	// "github.com/tchajed/goose/machine"
)

type Server struct {
	mu        *sync.Mutex
	epoch     uint64
	sealed    bool
	sm        *StateMachine
	nextIndex uint64

	isPrimary bool
	// clerks    []*Clerk

	durableIndex   uint64
	backgroundCond *sync.Cond // either entered a new epoch or made something new durable
	memlogIndex    uint64
	memlog         [][]byte

	acceptedIndex []uint64

	commitIndex uint64
	commitConds []*sync.Cond
}

func min(l []uint64) uint64 {
	var m = l[0]
	i := 1
	for i < len(l) {
		if l[i] < m {
			m = l[i]
		}
		i++
	}
	return m
}

// precondition: operations through opIndex have been accepted by server serverIndex
func (s *Server) handleNewAcceptedOp(epoch uint64, serverIndex uint64, opIndex uint64) {
	log.Printf("Handling new operation acceptance")
	s.mu.Lock()
	if s.epoch == epoch {
		prevIndex := s.acceptedIndex[serverIndex]
		if opIndex > prevIndex {
			s.acceptedIndex[serverIndex] = opIndex

			newCommitIndex := min(s.acceptedIndex)
			doneCommitConds := s.commitConds[:newCommitIndex-s.commitIndex]
			s.commitConds = s.commitConds[newCommitIndex-s.commitIndex:]
			s.commitIndex = newCommitIndex

			var i = uint64(0)
			for i < uint64(len(doneCommitConds)) {
				doneCommitConds[i].Signal()
				i++
			}
		}
	}
	s.mu.Unlock()
}

func (s *Server) backgroundThread(epoch uint64, i uint64, backupServer grove_ffi.Address) {
	clerk := MakeClerk(backupServer)
	log.Printf("Background thread made clerk")
	for {
		s.mu.Lock()
		log.Printf("Background thread got lock")
		for s.epoch == epoch && s.durableIndex <= s.memlogIndex {
			log.Printf("background thread waiting")
			s.backgroundCond.Wait()
		}
		if s.epoch != epoch {
			log.Printf("Background thread in epoch %d dying in epoch %d", epoch, s.epoch)
			s.mu.Unlock()
			break
		}
		// else, s.durableIndex > 0

		// log.Printf("background thread about to send RPC")
		op := s.memlog[0]
		index := s.memlogIndex
		s.memlogIndex = std.SumAssumeNoOverflow(s.memlogIndex, 1)
		s.memlog = s.memlog[1:]
		s.mu.Unlock()

		args := &ApplyAsBackupArgs{
			epoch: epoch,
			index: index,
			op:    op,
		}

		waitFn := clerk.StartApplyAsBackup(args)

		// go func() {
			err := waitFn()
			log.Printf("Get %+v", err)
			if err == e.None {
				s.handleNewAcceptedOp(epoch, i, index)
				continue
				// return
			}

			log.Printf("Looping")
			for {
				err := clerk.ApplyAsBackup(args)

				if err == e.None {
					break
				}
				if err == e.OutOfOrder || err == e.Timeout {
					continue
				} else {
					// FIXME: this should signal to the main thread that we're no
					// longer the primary.
					log.Fatalf("Got error %+v", err)
				}
			}
			s.handleNewAcceptedOp(epoch, i, index)
		// }()

	}
}

// called on the primary server to apply a new operation.
func (s *Server) Apply(op Op) *ApplyReply {
	reply := new(ApplyReply)
	reply.Reply = nil

	opCommitCond := sync.NewCond(s.mu)
	s.mu.Lock()
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

	// add to memlog
	s.memlog = append(s.memlog, op)
	s.commitConds = append(s.commitConds, opCommitCond)

	s.nextIndex = std.SumAssumeNoOverflow(s.nextIndex, 1)
	nextIndex := s.nextIndex
	s.mu.Unlock()

	waitForDurable()

	// increase durableIndex if someone else hasn't already increased it
	s.mu.Lock()
	if s.durableIndex < nextIndex {
		s.durableIndex = nextIndex
		// tell backup background threads that there's a new op they can send
		s.backgroundCond.Broadcast()
	}

	// wait for op to be committed or for us to no longer be leader
	for s.commitIndex < nextIndex {
		log.Printf("Waiting for op to be committed")
		opCommitCond.Wait()
	}
	s.mu.Unlock()
	log.Printf("Op committed")

	// log.Println("Apply() returned ", err)
	reply.Err = e.None
	return reply
}

// requires that we've already at least entered this epoch
// returns true iff stale
func (s *Server) isEpochStale(epoch uint64) bool {
	return s.epoch != epoch
}

var outOfOrder = uint64(0)
var backupApplies = uint64(0)

// called on backup servers to apply an operation so it is replicated and
// can be considered committed by primary.
func (s *Server) ApplyAsBackup(args *ApplyAsBackupArgs) e.Error {
	s.mu.Lock()
	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return e.Stale
	}
	if s.sealed {
		s.mu.Unlock()
		return e.Stale
	}

	backupApplies += 1
	ratio := float32(outOfOrder) / float32(backupApplies)
	if args.index != s.nextIndex {
		outOfOrder += 1
		log.Printf("Expected %d got %d", s.nextIndex, args.index)
		s.mu.Unlock()

		return e.OutOfOrder
	}

	// apply it locally
	_, waitFn := s.sm.StartApply(args.op)
	s.nextIndex += 1

	s.mu.Unlock()
	if machine.RandomUint64()%50000 == 1 {
		log.Println("Ratio of e.OutOfOrder:", ratio)
	}

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
	if args.Epoch < s.epoch {
		log.Println("Stale BecomePrimary request")
		s.mu.Unlock()
		return e.Stale
	}
	log.Println("Became Primary")
	s.isPrimary = true

	s.memlog = make([][]byte, 0)
	s.memlogIndex = s.nextIndex
	s.durableIndex = s.nextIndex
	s.commitIndex = s.nextIndex

	s.acceptedIndex = make([]uint64, len(args.Replicas) - 1)

	// XXX: should probably not bother doing this if we are already the primary
	// in this epoch
	var i = uint64(0)
	s.backgroundCond = sync.NewCond(s.mu)
	for i < uint64(len(args.Replicas) - 1) {
		go s.backgroundThread(args.Epoch, i, args.Replicas[i+1])
		i++
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
