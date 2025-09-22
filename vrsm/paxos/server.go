package paxos

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/asyncfile"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/mit-pdos/gokv/vrsm/paxos/applyasfollowerargs_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/applyasfollowerreply_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/enternewepochargs_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/enternewepochreply_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/error_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos/paxosstate_gk"
)

type Server struct {
	mu      *sync.Mutex
	ps      *paxosstate_gk.S
	storage *asyncfile.AsyncFile
	clerks  []*singleClerk
}

func (s *Server) withLock(f func(ps *paxosstate_gk.S)) {
	s.mu.Lock()
	f(s.ps)
	waitFn := s.storage.Write(paxosstate_gk.Marshal(make([]byte, 0), *s.ps))
	s.mu.Unlock()
	waitFn()
}

func (s *Server) applyAsFollower(args *applyasfollowerargs_gk.S, reply *applyasfollowerreply_gk.S) {
	s.withLock(func(ps *paxosstate_gk.S) {
		if ps.Epoch <= args.Epoch {
			if ps.AcceptedEpoch == args.Epoch {
				if ps.NextIndex < args.NextIndex {
					ps.NextIndex = args.NextIndex
					ps.State = args.State
					reply.Err = error_gk.ENone
				} else { // args.nextIndex < s.nextIndex
					reply.Err = error_gk.ENone
				}
			} else { // s.acceptedEpoch < args.epoch, because s.acceptedEpoch <= s.epoch <= args.epoch
				ps.AcceptedEpoch = args.Epoch
				ps.Epoch = args.Epoch
				ps.State = args.State
				ps.NextIndex = args.NextIndex
				ps.IsLeader = false
				reply.Err = error_gk.ENone
			}
		} else {
			reply.Err = error_gk.EEpochStale
		}
	})
}

// NOTE:
// This will vote yes only the first time it's called in an epoch.
// If you have too aggressive of a timeout and end up retrying this, the retry
// might fail because it may be the second execution of enterNewEpoch(epoch) on
// the server.
// Solution: either conservative (maybe double) timeouts, or don't use this for
// leader election, only for coming up with a valid proposal.
func (s *Server) enterNewEpoch(args *enternewepochargs_gk.S, reply *enternewepochreply_gk.S) {
	s.withLock(func(ps *paxosstate_gk.S) {
		if ps.Epoch >= args.Epoch {
			reply.Err = error_gk.EEpochStale
			return
		}
		// else, s.epoch < args.epoch
		ps.IsLeader = false
		ps.Epoch = args.Epoch
		reply.AcceptedEpoch = ps.AcceptedEpoch
		reply.NextIndex = ps.NextIndex
		reply.State = ps.State
	})
}

func (s *Server) TryBecomeLeader() {
	log.Println("started trybecomeleader")
	// defer log.Println("finished trybecomeleader")
	s.mu.Lock()
	if s.ps.IsLeader {
		log.Println("already leader")
		s.mu.Unlock()
		return
	}
	// pick a new epoch number
	clerks := s.clerks
	args := &enternewepochargs_gk.S{Epoch: s.ps.Epoch + 1}
	s.mu.Unlock()

	var numReplies = uint64(0)
	replies := make([]*enternewepochreply_gk.S, uint64(len(clerks)))

	mu := new(sync.Mutex)
	numReplies_cond := sync.NewCond(mu)
	n := uint64(len(clerks))

	for i, ck := range clerks {
		go func() {
			reply := ck.enterNewEpoch(args)
			mu.Lock()
			numReplies += 1
			replies[i] = reply
			if 2*numReplies > n {
				numReplies_cond.Signal()
			}
			mu.Unlock()
		}()
	}

	mu.Lock()
	// wait for a quorum of replies
	for 2*numReplies <= n {
		numReplies_cond.Wait()
	}

	var latestReply *enternewepochreply_gk.S
	var numSuccesses = uint64(0)
	for _, reply := range replies {
		if reply != nil {
			if reply.Err == error_gk.ENone {
				if numSuccesses == 0 {
					latestReply = reply
				} else {
					if latestReply.AcceptedEpoch < reply.AcceptedEpoch {
						latestReply = reply
					} else if latestReply.AcceptedEpoch == reply.AcceptedEpoch &&
						reply.NextIndex > latestReply.NextIndex {
						latestReply = reply
					}
				}
				numSuccesses += 1
			}
		}
	}

	if 2*numSuccesses > n {
		// RULE: lock s.mu after mu
		// XXX: withLock has a disk write inside of it, so `mu` will be held for
		// a long time here. This is ok because it only blocks the late RPC
		// replies from replica servers, which we anyways won't look at.
		s.withLock(func(ps *paxosstate_gk.S) {
			if ps.Epoch <= args.Epoch {
				log.Printf("succeeded becomeleader in epoch %d\n", args.Epoch)
				ps.Epoch = args.Epoch
				ps.IsLeader = true
				ps.AcceptedEpoch = ps.Epoch
				ps.NextIndex = latestReply.NextIndex
				ps.State = latestReply.State
			}
		})
		mu.Unlock()
	} else {
		mu.Unlock()
		log.Println("failed becomeleader")
	}
}

func (s *Server) TryAcquire() (error_gk.E, *[]byte, func() error_gk.E) {
	var retErr error_gk.E

	s.mu.Lock()
	if !s.ps.IsLeader {
		s.mu.Unlock()
		var n *[]byte // XXX: hack for Goose; want to just return nil pointer,
		// but Goose translates that into a nil slice.
		return error_gk.ENotLeader, n, nil
	}

	// between the previous lines of code and the invocation of tryRelease, the user is allowed to
	// modify the state however they wish.

	tryRelease := func() error_gk.E {
		s.ps.NextIndex = std.SumAssumeNoOverflow(s.ps.NextIndex, 1)
		args := &applyasfollowerargs_gk.S{Epoch: s.ps.Epoch, NextIndex: s.ps.NextIndex, State: s.ps.State}
		waitFn := s.storage.Write(paxosstate_gk.Marshal(make([]byte, 0), *s.ps))
		s.mu.Unlock()
		waitFn()

		clerks := s.clerks

		var numReplies = uint64(0)
		replies := make([]*applyasfollowerreply_gk.S, uint64(len(clerks)))
		mu := new(sync.Mutex)
		numReplies_cond := sync.NewCond(mu)
		n := uint64(len(clerks))

		for i, ck := range clerks {
			ck := ck
			i := i
			go func() {
				reply := ck.applyAsFollower(args)

				mu.Lock()
				numReplies += 1
				replies[i] = reply
				if 2*numReplies > n {
					numReplies_cond.Signal()
				}
				mu.Unlock()
			}()
		}

		mu.Lock()
		// wait for a quorum of replies
		for 2*numReplies <= n {
			numReplies_cond.Wait()
		}

		var numSuccesses = uint64(0)
		for _, reply := range replies {
			if reply != nil {
				if reply.Err == error_gk.ENone {
					numSuccesses += 1
				}
			}
		}

		if 2*numSuccesses > n {
			retErr = error_gk.ENone
		} else {
			retErr = error_gk.EEpochStale
		}
		return retErr
	}
	return error_gk.ENone, &s.ps.State, tryRelease
}

func (s *Server) WeakRead() []byte {
	s.mu.Lock()
	ret := s.ps.State
	s.mu.Unlock()
	return ret
}

func makeServer(fname string, initstate []byte, config []grove_ffi.Address) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)

	s.clerks = make([]*singleClerk, 0)
	for _, host := range config {
		s.clerks = append(s.clerks, MakeSingleClerk(host))
	}

	var encstate []byte
	encstate, s.storage = asyncfile.MakeAsyncFile(fname)
	if len(encstate) == 0 {
		s.ps = new(paxosstate_gk.S)
		s.ps.State = initstate
	} else {
		(*s.ps), _ = paxosstate_gk.Unmarshal(encstate)
	}
	return s
}

func StartServer(fname string, initstate []byte, me grove_ffi.Address, config []grove_ffi.Address) *Server {
	s := makeServer(fname, initstate, config)

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_APPLY_AS_FOLLOWER] = func(raw_args []byte, raw_reply *[]byte) {
		reply := new(applyasfollowerreply_gk.S)
		args, _ := applyasfollowerargs_gk.Unmarshal(raw_args)
		s.applyAsFollower(&args, reply)
		*raw_reply = applyasfollowerreply_gk.Marshal(make([]byte, 0), *reply)
	}

	handlers[RPC_ENTER_NEW_EPOCH] = func(raw_args []byte, raw_reply *[]byte) {
		reply := new(enternewepochreply_gk.S)
		args, _ := enternewepochargs_gk.Unmarshal(raw_args)
		s.enterNewEpoch(&args, reply)
		*raw_reply = enternewepochreply_gk.Marshal(make([]byte, 0), *reply)
	}

	handlers[RPC_BECOME_LEADER] = func(raw_args []byte, raw_reply *[]byte) {
		s.TryBecomeLeader()
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
	return s
}
