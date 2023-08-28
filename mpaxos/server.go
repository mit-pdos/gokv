package mpaxos

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/asyncfile"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type paxosState struct {
	epoch         uint64
	acceptedEpoch uint64
	nextIndex     uint64
	state         []byte
	isLeader      bool
}

type Server struct {
	mu      *sync.Mutex
	ps      *paxosState
	storage *asyncfile.File
	clerks  []*singleClerk
}

func (s *Server) withLock(f func(ps *paxosState)) {
	s.mu.Lock()
	f(s.ps)
	waitFn := s.storage.AsyncWrite(encodePaxosState(s.ps))
	s.mu.Unlock()
	waitFn()
}

func (s *Server) applyAsFollower(args *applyAsFollowerArgs, reply *applyAsFollowerReply) {
	s.withLock(func(ps *paxosState) {
		if ps.epoch <= args.epoch {
			ps.isLeader = false
			if ps.acceptedEpoch == args.epoch {
				if ps.nextIndex <= args.nextIndex {
					ps.nextIndex = args.nextIndex + 1
					ps.state = args.state
					reply.err = ENone
				} else { // args.nextIndex < s.nextIndex
					reply.err = ENone
				}
			} else { // s.acceptedEpoch < args.epoch, because s.acceptedEpoch <= s.epoch <= args.epoch
				ps.acceptedEpoch = args.epoch
				ps.epoch = args.epoch
				ps.state = args.state
				ps.nextIndex = args.nextIndex
				reply.err = ENone
			}
		} else {
			reply.err = EEpochStale
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
func (s *Server) enterNewEpoch(args *enterNewEpochArgs, reply *enterNewEpochReply) {
	s.withLock(func(ps *paxosState) {
		if ps.epoch >= args.epoch {
			reply.err = EEpochStale
			return
		}
		// else, s.epoch < args.epoch
		ps.isLeader = false
		ps.epoch = args.epoch
		reply.acceptedEpoch = ps.acceptedEpoch
		reply.nextIndex = ps.nextIndex
		reply.state = ps.state
		s.mu.Unlock()
	})
}

func (s *Server) becomeLeader() {
	log.Println("started trybecomeleader")
	// defer log.Println("finished trybecomeleader")
	s.mu.Lock()
	if s.ps.isLeader {
		log.Println("already leader")
		s.mu.Unlock()
		return
	}
	// pick a new epoch number
	clerks := s.clerks
	args := &enterNewEpochArgs{epoch: s.ps.epoch + 1}
	s.mu.Unlock()

	var numReplies = uint64(0)
	replies := make([]*enterNewEpochReply, uint64(len(clerks)))

	mu := new(sync.Mutex)
	numReplies_cond := sync.NewCond(mu)
	n := uint64(len(clerks))

	for i, ck := range clerks {
		ck := ck
		i := i
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

	var latestReply *enterNewEpochReply
	var numSuccesses = uint64(0)
	for _, reply := range replies {
		if reply != nil {
			if reply.err == ENone {
				if numSuccesses == 0 {
					latestReply = reply
				} else {
					if latestReply.acceptedEpoch < reply.acceptedEpoch {
						latestReply = reply
					} else if latestReply.acceptedEpoch == reply.acceptedEpoch &&
						reply.nextIndex > latestReply.nextIndex {
						latestReply = reply
					}
				}
				numSuccesses += 1
			}
		}
	}

	if 2*numSuccesses > n {
		log.Printf("succeeded becomeleader in epoch %d\n", args.epoch)
		// RULE: lock s.mu after mu
		// XXX: withLock has a disk write inside of it, so `mu` will be held for
		// a long time here. This is ok because it only blocks the late RPC
		// replies from replica servers, which we anyways won't look at.
		s.withLock(func(ps *paxosState) {
			if ps.epoch < args.epoch {
				ps.epoch = args.epoch
				ps.isLeader = true
				ps.acceptedEpoch = ps.epoch
				ps.nextIndex = latestReply.nextIndex
				ps.state = latestReply.state
			}
		})
		mu.Unlock()
	} else {
		mu.Unlock()
		log.Println("failed becomeleader")
	}
}

func (s *Server) apply(applyFn func([]byte) ([]byte, []byte)) (Error, []byte) {
	var retErr Error
	var retVal []byte
	var args *applyAsFollowerArgs

	// make proposal
	s.withLock(func(ps *paxosState) {
		if !ps.isLeader {
			retErr = ENotLeader
			return
		}
		ps.state, retVal = applyFn(ps.state)
		args = &applyAsFollowerArgs{epoch: ps.epoch, nextIndex: ps.nextIndex, state: ps.state}
		ps.nextIndex = std.SumAssumeNoOverflow(ps.nextIndex, 1)
	})
	clerks := s.clerks
	if retErr != 0 {
		return retErr, nil
	}

	var numReplies = uint64(0)
	replies := make([]*applyAsFollowerReply, uint64(len(clerks)))
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
			if reply.err == ENone {
				numSuccesses += 1
			}
		}
	}

	if 2*numSuccesses > n {
		retErr = ENone
	} else {
		retErr = EEpochStale
	}
	return retErr, retVal
}

func makeServer(fname string, applyFn func([]byte, []byte) ([]byte, []byte),
	config []grove_ffi.Address) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)

	var encstate []byte
	encstate, s.storage = asyncfile.MakeFile(fname)
	s.ps = decodePaxosState(encstate)

	s.clerks = make([]*singleClerk, len(config))
	n := uint64(len(s.clerks))
	var i = uint64(0)
	for i < n {
		s.clerks[i] = makeSingleClerk(config[i])
		i += 1
	}
	return s
}

func StartServer(fname string, me grove_ffi.Address,
	applyFn func([]byte, []byte) ([]byte, []byte),
	config []grove_ffi.Address) {
	s := makeServer(fname, applyFn, config)

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_APPLY_AS_FOLLOWER] = func(raw_args []byte, raw_reply *[]byte) {
		reply := new(applyAsFollowerReply)
		args := decodeApplyAsFollowerArgs(raw_args)
		s.applyAsFollower(args, reply)
		*raw_reply = encodeApplyAsFollowerReply(reply)
	}

	handlers[RPC_ENTER_NEW_EPOCH] = func(raw_args []byte, raw_reply *[]byte) {
		reply := new(enterNewEpochReply)
		args := decodeEnterNewEpochArgs(raw_args)
		s.enterNewEpoch(args, reply)
		*raw_reply = encodeEnterNewEpochReply(reply)
	}

	handlers[RPC_BECOME_LEADER] = func(raw_args []byte, raw_reply *[]byte) {
		s.becomeLeader()
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
}
