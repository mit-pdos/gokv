package reconf

import (
	"log"
	"sync"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/paxi/reconf/config_gk"
	"github.com/mit-pdos/gokv/paxi/reconf/error_gk"
	"github.com/mit-pdos/gokv/paxi/reconf/monotonicvalue_gk"
	"github.com/mit-pdos/gokv/paxi/reconf/preparereply_gk"
	"github.com/mit-pdos/gokv/paxi/reconf/proposeargs_gk"
	"github.com/mit-pdos/gokv/paxi/reconf/trycommitreply_gk"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
	// "github.com/davecgh/go-spew/spew"
)

func MonotonicValGreaterThan(lhs *monotonicvalue_gk.S, rhs *monotonicvalue_gk.S) bool {
	return lhs.Version > rhs.Version
}

type Replica struct {
	mu           *sync.Mutex
	promisedTerm uint64

	acceptedTerm uint64
	acceptedMVal *monotonicvalue_gk.S

	clerkPool *ClerkPool

	isLeader bool
	// Leader state
	// acceptedVersions map[grove_ffi.Address]uint64
}

func (r *Replica) PrepareRPC(term uint64, reply *preparereply_gk.S) {
	r.mu.Lock()
	if term > r.promisedTerm {
		r.promisedTerm = term
		reply.Term = r.acceptedTerm
		reply.Val = *r.acceptedMVal
		reply.Err = error_gk.ENone
	} else {
		reply.Err = error_gk.ETermStale
		reply.Val = monotonicvalue_gk.S{}
		reply.Val.Conf = config_gk.S{}
		reply.Term = r.promisedTerm
	}
	r.mu.Unlock()
}

func (r *Replica) ProposeRPC(term uint64, v *monotonicvalue_gk.S) error_gk.E {
	r.mu.Lock()
	if term >= r.promisedTerm {
		r.promisedTerm = term
		r.acceptedTerm = term
		if MonotonicValGreaterThan(v, r.acceptedMVal) {
			r.acceptedMVal = v
		}
		r.mu.Unlock()
		return error_gk.ENone
	} else {
		r.mu.Unlock()
		return error_gk.ETermStale
	}
}

func (r *Replica) TryBecomeLeader() bool {
	r.mu.Lock()
	newTerm := r.promisedTerm + 1 // don't need to bother incrementing; will invoke RPC on ourselves
	r.promisedTerm = newTerm

	var highestTerm uint64
	highestTerm = 0
	var highestVal *monotonicvalue_gk.S
	highestVal = r.acceptedMVal // if no one in our majority has accepted a value, we'll propose this one
	conf := r.acceptedMVal.Conf
	r.mu.Unlock()

	mu := new(sync.Mutex)

	prepared := make(map[grove_ffi.Address]bool)

	ForEachConfigMember(&conf, func(addr grove_ffi.Address) {
		go func() {
			reply_ptr := new(preparereply_gk.S)
			r.clerkPool.PrepareRPC(addr, newTerm, reply_ptr)

			if reply_ptr.Err == error_gk.ENone {
				mu.Lock()
				prepared[addr] = true

				if reply_ptr.Term > highestTerm {
					highestVal = &reply_ptr.Val
				} else if reply_ptr.Term == highestTerm {
					if MonotonicValGreaterThan(highestVal, &reply_ptr.Val) {
						highestVal = &reply_ptr.Val
					}
				}
				mu.Unlock()
			} else {
				// If we did the following, then whenever a single other node
				// has a higher term number than us, we would just give up on
				// our attempt at becoming leader. We should only do this if we
				// fail to get a quorum.
				//
				// if reply_ptr.Term > r.promisedTerm {
				// r.promisedTerm = reply_ptr.Term
				// }
			}
		}()
	})

	// FIXME: put this in a condvar loop with timeout
	primitive.Sleep(50 * 1_000_000) // 50 ms
	mu.Lock()
	if IsQuorum(&(highestVal.Conf), prepared) {
		// We successfully became the leader
		r.mu.Lock()
		if r.promisedTerm == newTerm {
			r.acceptedMVal = highestVal
			r.isLeader = true
			// r.acceptedIndices = make(map[grove_ffi.Address]uint64)
		}
		r.mu.Unlock()
		mu.Unlock()
		return true
	}
	mu.Unlock()
	return false
}

// Returns true iff there was an error;
// The error is either that r is not currently a primary, or that r was unable
// to commit the value within one round of commits.
//
// mvalModifier is not allowed to modify the version number in the given mval.
func (r *Replica) tryCommit(mvalModifier func(*monotonicvalue_gk.S), reply *trycommitreply_gk.S) {
	r.mu.Lock()
	if !r.isLeader {
		r.mu.Unlock()
		reply.Err = error_gk.ENotLeader
		return
	}
	mvalModifier(r.acceptedMVal)
	// if !r.acceptedMVal.conf.Contains(r.me) {
	// I don't think we need to do this; even if the new configuration we're
	// going to doesn't contain us, we're still the leader of this term. In
	// fact, I think we can even commit new entries even if we're not
	// actually in the config!
	// r.isLeader = false
	// }

	log.Printf("Trying to commit value; node state: %+v\n", r)
	// spew.Printf("MVal state: %+v\n", r.acceptedMVal)

	r.acceptedMVal.Version += 1
	term := r.promisedTerm
	mval := r.acceptedMVal
	r.mu.Unlock()

	mu := new(sync.Mutex)
	accepted := make(map[grove_ffi.Address]bool)

	ForEachConfigMember(&mval.Conf, func(addr grove_ffi.Address) {
		go func() {
			if r.clerkPool.ProposeRPC(addr, term, mval) {
				mu.Lock()
				accepted[addr] = true
				mu.Unlock()
			}
		}()
	})

	// FIXME: put this in a condvar loop with timeout
	primitive.Sleep(100 * 1_000_000) // 100ms
	mu.Lock()
	if IsQuorum(&mval.Conf, accepted) {
		reply.Err = error_gk.ENone
		reply.Version = mval.Version
	} else {
		reply.Err = error_gk.EQuorumFailed
	}
	log.Printf("Result of trying to commit: %+v\n", reply)
}

func (r *Replica) TryCommitVal(v []byte, reply *trycommitreply_gk.S) {
	r.mu.Lock()
	if !r.isLeader {
		r.mu.Unlock()
		r.TryBecomeLeader()
	} else {
		r.mu.Unlock()
	}

	r.tryCommit(func(mval *monotonicvalue_gk.S) {
		mval.Val = v
	}, reply)
}

// requires that newConfig has overlapping quorums with r.config
func (r *Replica) TryEnterNewConfig(newMembers []grove_ffi.Address) {
	reply := new(trycommitreply_gk.S)
	r.tryCommit(func(mval *monotonicvalue_gk.S) {
		if len(mval.Conf.NextMembers) == 0 {
			mval.Conf.NextMembers = newMembers
		}
	}, reply)

	r.tryCommit(func(mval *monotonicvalue_gk.S) {
		if len(mval.Conf.NextMembers) != 0 {
			mval.Conf.Members = mval.Conf.NextMembers
			mval.Conf.NextMembers = make([]grove_ffi.Address, 0)
		}
	}, reply)
}

func StartReplicaServer(me grove_ffi.Address, initConfig *config_gk.S) {
	s := new(Replica)

	s.mu = new(sync.Mutex)
	s.promisedTerm = 0
	s.acceptedTerm = 0
	s.acceptedMVal = new(monotonicvalue_gk.S)
	s.acceptedMVal.Conf = *initConfig

	s.clerkPool = MakeClerkPool()
	s.isLeader = false

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_PREPARE] = func(args []byte, raw_reply *[]byte) {
		term, _ := marshal.ReadInt(args)
		reply := new(preparereply_gk.S)
		s.PrepareRPC(term, reply)
		*raw_reply = preparereply_gk.Marshal(make([]byte, 0), *reply)
		// XXX: Why is this here?
		preparereply_gk.Unmarshal(*raw_reply)
	}

	handlers[RPC_PROPOSE] = func(raw_args []byte, raw_reply *[]byte) {
		args, _ := proposeargs_gk.Unmarshal(raw_args)
		reply := s.ProposeRPC(args.Term, &args.Val)
		*raw_reply = error_gk.Marshal(make([]byte, 0, 8), reply)
	}

	handlers[RPC_TRY_COMMIT_VAL] = func(raw_args []byte, raw_reply *[]byte) {
		log.Println("RPC_TRY_COMMIT_VAL")
		val := raw_args
		reply := new(trycommitreply_gk.S)
		s.TryCommitVal(val, reply)
		*raw_reply = error_gk.Marshal(make([]byte, 0, 8), reply.Err)
	}

	handlers[RPC_TRY_CONFIG_CHANGE] = func(raw_args []byte, raw_reply *[]byte) {
		args, _ := marshal.ReadSliceLenPrefix(raw_args, marshal.ReadInt)
		s.TryEnterNewConfig(args)
		*raw_reply = make([]byte, 0)
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
}
