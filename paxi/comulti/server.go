package comulti

// Fault tolerant shared log library.
//
// This generalizes single-slot paxos by making the value be the full log.
// Instead of getting knowledge on what value the single-slot paxos protocol has
// selected, you only get knowledge that the value agreed upon by paxos _at
// least_ contains some prefix of a log.
//
// If you stare at this for long enough, you'll see that it's similar to Raft.
// In fact, if you start optimizing this some more (so e.g. you don't need to
// send the *whole* log on every request), you'll end up converging to something
// much closer to Raft.
//
// One could also use this idea to implement a fault-tolerant monotonic counter
// without worrying about any sort of log.

import (
	"time"
	"sync"
)

type Entry = uint64

type Replica struct {
	mu         *sync.Mutex
	promisedPN uint64 // server has promised not to accept proposals below this

	logPN uint64  // proposal number of accepted val
	log   []Entry // the value itself

	commitIndex   uint64

	peers []*Clerk

	isLeader bool // this means that we own the proposal with number logPN
	leaderCond *sync.Cond
	commitCond *sync.Cond
	acceptedIndex []uint64 // how much of the log has each peer accepted?
}

type PrepareReply struct {
	Success bool
	Log     []Entry // full log;
	Pn      uint64
}

func (r *Replica) PrepareRPC(pn uint64, reply *PrepareReply) {
	r.mu.Lock()
	if pn > r.promisedPN {
		r.promisedPN = pn

		reply.Pn = r.logPN
		reply.Log = r.log
		reply.Success = true
	} else {
		reply.Success = false
		reply.Pn = r.promisedPN
	}
	r.mu.Unlock()
}

type ProposeArgs struct {
	Pn          uint64
	CommitIndex uint64
	Log         []Entry
}

func (r *Replica) ProposeRPC(pn uint64, commitIndex uint64, val []Entry) bool {
	r.mu.Lock()
	if pn >= r.promisedPN && pn >= r.logPN {
		if pn > r.logPN {
			r.log = val
			r.logPN = pn
		} else if len(val) > len(r.log) {
			r.log = val
		}
		if commitIndex > r.commitIndex {
			r.commitIndex = commitIndex
		}
		// What if r.commitIndex > commitIndex?  pn |-> val means that val
		// contains anything that might have been committed before. That means
		// that commitIndex >= r.commitIndex must always be true since pn >=
		// r.logPN.
		r.mu.Unlock()
		return true
	} else {
		r.mu.Unlock()
		return false
	}
}

func (r *Replica) Start(cmd Entry) bool {
	r.mu.Lock()
	if r.isLeader {
		r.log = append(r.log, cmd)
	}
	r.mu.Unlock()
	return true
}

func (r *Replica) GetLog() []Entry {
	return nil
}

// proposes whatever is in our log
// requires that the lock is held and that we are the leader
func (r *Replica) doPropose(peerIdx uint64) {
	pn := r.logPN
	log := r.log
	r.mu.Unlock()
	a := r.peers[peerIdx].Propose(pn, log)
	if a {
		r.mu.Lock()
		if r.logPN == pn && uint64(len(log)) > r.acceptedIndex[peerIdx] {
			r.acceptedIndex[peerIdx] = uint64(len(log))
		}
		r.mu.Unlock()
	}
}

func (r *Replica) commitThread() {
	for {
		for !r.isLeader {
			r.leaderCond.Wait()
		}

		// update commitIndex
		for r.isLeader {
			// FIXME: increase commitIndex if possible
			r.commitCond.Wait()
		}

	}
}

func (r *Replica) replicaThread(i uint64) {
	r.mu.Lock()
	for {
		for !r.isLeader {
			r.leaderCond.Wait()
		}

		r.doPropose(i)
		time.Sleep(10 * time.Millisecond)
	}
}

// returns true iff there was an error
func (r *Replica) TryDecide() bool {
	r.mu.Lock()
	pn := r.promisedPN + 1 // don't need to bother incrementing; will invoke RPC on ourselves
	r.mu.Unlock()

	var numPrepared uint64
	numPrepared = 0
	var highestPn uint64
	highestPn = 0
	var highestVal ValType
	highestVal = v // if no one in our majority has accepted a value, we'll propose this one
	mu := new(sync.Mutex)

	for _, peer := range r.peers { // XXX: peers is readonly
		local_peer := peer
		go func() {
			reply_ptr := new(PrepareReply)
			local_peer.Prepare(pn, reply_ptr) // TODO: replace with real RPC

			if reply_ptr.Success {
				mu.Lock()
				numPrepared = numPrepared + 1
				if reply_ptr.Pn > highestPn {
					highestVal = reply_ptr.Val
					highestPn = reply_ptr.Pn
				}
				mu.Unlock()
			}
		}()
	}

	// FIXME: put this in a condvar loop with timeout
	mu.Lock()
	n := numPrepared
	proposeVal := highestVal
	mu.Unlock()

	if 2*n > uint64(len(r.peers)) {
		mu2 := new(sync.Mutex)
		var numAccepted uint64
		numAccepted = 0

		for _, peer := range r.peers {
			local_peer := peer
			// each thread talks to a unique peer
			go func() {
				r := local_peer.Propose(pn, proposeVal) // TODO: replace with real RPC
				if r {
					mu2.Lock()
					numAccepted = numAccepted + 1
					mu2.Unlock()
				}
			}()
		}

		mu2.Lock()
		n := numAccepted
		mu2.Unlock()

		if 2*n > uint64(len(r.peers)) {
			*outv = proposeVal
			return false
		} else {
			return true
		}
	} else {
		return true
	}
}
