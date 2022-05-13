package reconf

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"log"
	"sync"
)

func (lhs *MonotonicValue) GreaterThan(rhs *MonotonicValue) bool {
	return lhs.version > rhs.version
}

type Replica struct {
	mu           *sync.Mutex
	promisedTerm uint64

	acceptedTerm uint64
	acceptedMVal *MonotonicValue

	clerkPool *ClerkPool

	isLeader bool
	// Leader state
	// acceptedVersions map[grove_ffi.Address]uint64
}

func (r *Replica) PrepareRPC(term uint64, reply *PrepareReply) {
	r.mu.Lock()
	if term > r.promisedTerm {
		r.promisedTerm = term
		reply.Term = r.acceptedTerm
		reply.Val = r.acceptedMVal
		reply.Success = true
	} else {
		reply.Success = false
		reply.Term = r.promisedTerm
	}
	r.mu.Unlock()
}

func (r *Replica) ProposeRPC(term uint64, v *MonotonicValue) bool {
	r.mu.Lock()
	if term >= r.promisedTerm {
		r.promisedTerm = term
		r.acceptedTerm = term
		if v.GreaterThan(r.acceptedMVal) {
			r.acceptedMVal = v
		}
		r.mu.Unlock()
		return true
	} else {
		r.mu.Unlock()
		return false
	}
}

func (r *Replica) TryBecomeLeader() bool {
	r.mu.Lock()
	newTerm := r.promisedTerm + 1 // don't need to bother incrementing; will invoke RPC on ourselves
	r.promisedTerm = newTerm

	var highestTerm uint64
	highestTerm = 0
	var highestVal *MonotonicValue
	highestVal = r.acceptedMVal // if no one in our majority has accepted a value, we'll propose this one
	conf := r.acceptedMVal.conf
	r.mu.Unlock()

	mu := new(sync.Mutex)

	prepared := make(map[grove_ffi.Address]bool)

	conf.ForEachMember(func(addr grove_ffi.Address) {
		go func() {
			reply_ptr := new(PrepareReply)
			r.clerkPool.PrepareRPC(addr, newTerm, reply_ptr) // TODO: replace with real RPC

			if reply_ptr.Success {
				mu.Lock()
				prepared[addr] = true

				if reply_ptr.Term > highestTerm {
					highestVal = reply_ptr.Val
				} else if reply_ptr.Term == highestTerm {
					if highestVal.GreaterThan(reply_ptr.Val) {
						highestVal = reply_ptr.Val
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

	grove_ffi.Sleep(50 * 1_000_000) // 50 ms

	// FIXME: put this in a condvar loop with timeout
	mu.Lock()
	if IsQuorum(highestVal.conf, prepared) {
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

type TryCommitReply struct {
	err     uint64
	version uint64
}

const (
	ENone         = uint64(0)
	ENotLeader    = uint64(1)
	EQuorumFailed = uint64(2)
)

// Returns true iff there was an error;
// The error is either that r is not currently a primary, or that r was unable
// to commit the value within one round of commits.
//
// mvalModifier is not allowed to modify the version number in the given mval.
func (r *Replica) tryCommit(mvalModifier func(*MonotonicValue), reply *TryCommitReply) {
	r.mu.Lock()
	if !r.isLeader {
		r.mu.Unlock()
		reply.err = ENotLeader
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

	r.acceptedMVal.version += 1
	term := r.promisedTerm
	mval := r.acceptedMVal
	r.mu.Unlock()

	mu := new(sync.Mutex)
	accepted := make(map[grove_ffi.Address]bool)

	mval.conf.ForEachMember(func(addr grove_ffi.Address) {
		go func() {
			if r.clerkPool.ProposeRPC(addr, term, mval) {
				mu.Lock()
				accepted[addr] = true
				mu.Unlock()
			}
		}()
	})

	grove_ffi.Sleep(100 * 1_000_000) // 100ms
	mu.Lock()
	if IsQuorum(mval.conf, accepted) {
		reply.err = ENone
		reply.version = mval.version
	} else {
		reply.err = EQuorumFailed
	}
	log.Printf("Result of trying to commit: %+v\n", reply)
}

func (r *Replica) TryCommitVal(v []byte, reply *TryCommitReply) {
	r.tryCommit(func(mval *MonotonicValue) {
		mval.val = v
	}, reply)
}

// requires that newConfig has overlapping quorums with r.config
func (r *Replica) TryEnterNewConfig(newMembers []grove_ffi.Address) {
	reply := new(TryCommitReply)
	r.tryCommit(func(mval *MonotonicValue) {
		if len(mval.conf.nextMembers) == 0 {
			mval.conf.nextMembers = newMembers
		}
	}, reply)

	r.tryCommit(func(mval *MonotonicValue) {
		if len(mval.conf.nextMembers) != 0 {
			mval.conf.members = mval.conf.nextMembers
			mval.conf.nextMembers = make([]grove_ffi.Address, 0)
		}
	}, reply)
}

func StartReplicaServer(me grove_ffi.Address, initConfig *Config) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_PREPARE] = func(args []byte, reply *[]byte) {
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
}
