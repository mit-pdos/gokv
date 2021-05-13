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
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
	"time"
)

type Entry = uint64

type Replica struct {
	mu         *sync.Mutex
	promisedPN uint64 // server has promised not to accept proposals below this

	logPN uint64  // proposal number of accepted val
	log   []Entry // the value itself

	commitIndex uint64

	peers []*Clerk

	isLeader      bool // this means that we own the proposal with number logPN
	leaderCond    *sync.Cond
	commitCond    *sync.Cond
	applyCond     *sync.Cond
	acceptedIndex []uint64 // how much of the log has each peer accepted?
	commitf       func(Entry)
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
			r.applyCond.Signal()
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

func (r *Replica) TryAppendRPC(cmd Entry) bool {
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
	commitIndex := r.commitIndex
	r.mu.Unlock()
	a := r.peers[peerIdx].Propose(pn, commitIndex, log)
	if a {
		r.mu.Lock()
		if r.logPN == pn && uint64(len(log)) > r.acceptedIndex[peerIdx] {
			r.acceptedIndex[peerIdx] = uint64(len(log))
			r.commitCond.Signal()
		}
		r.mu.Unlock()
	}
	r.mu.Lock()
}

func (r *Replica) applyThread() {
	var lastApplied uint64 // := 0

	r.mu.Lock()
	for {
		// fmt.Printf("lastApplied %+v\n", lastApplied)
		for r.commitIndex <= lastApplied {
			r.applyCond.Wait()
		}
		for lastApplied < r.commitIndex {
			r.commitf(r.log[lastApplied])
			lastApplied++
		}
	}
}

func (r *Replica) commitThread() {
	r.mu.Lock()
	for {
		for !r.isLeader {
			r.leaderCond.Wait()
		}

		// update commitIndex
		for r.isLeader {
			oldCommitIndex := r.commitIndex
			for { // increase commitIndex as much as possible
				tally := 0
				for _, a := range r.acceptedIndex {
					if a > r.commitIndex {
						tally++
					}
				}
				if 2*tally > len(r.peers) {
					r.commitIndex++
				} else {
					break
				}
			}
			// apply everything in the range [oldCommitIndex + 1, commitIndex]
			if r.commitIndex > oldCommitIndex {
				r.applyCond.Signal()
			}
			r.commitCond.Wait() // going to have to wait in any case, since we already committed as much as possible
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
		// fmt.Printf("%+v\n", r)
		time.Sleep(10 * time.Millisecond)
	}
}

// returns true iff there was an error
func (r *Replica) TryBecomeLeader() {
	r.mu.Lock()
	pn := r.promisedPN + 1 // don't need to bother incrementing; will invoke RPC on ourselves
	r.mu.Unlock()

	var numPrepared uint64
	numPrepared = 0
	var highestPn uint64
	highestPn = 0
	var highestLog []Entry
	highestLog = r.log // if no one in our majority has accepted a value, we'll propose this one
	mu := new(sync.Mutex)

	for _, peer := range r.peers { // XXX: peers is readonly
		local_peer := peer
		go func() {
			reply_ptr := new(PrepareReply)
			local_peer.Prepare(pn, reply_ptr)

			if reply_ptr.Success {
				mu.Lock()
				numPrepared = numPrepared + 1
				if reply_ptr.Pn > highestPn {
					highestLog = reply_ptr.Log
					highestPn = reply_ptr.Pn
				}
				mu.Unlock()
			}
		}()
	}

	// FIXME: put this in a condvar loop with timeout
	mu.Lock()
	if 2*numPrepared > uint64(len(r.peers)) && r.promisedPN <= pn && r.logPN <= pn {
		r.log = highestLog
		r.logPN = pn
		r.isLeader = true
	}
	mu.Unlock()
}

func MakeReplica(me uint64, commitf func(Entry), peerHosts []uint64, isLeader bool) *Replica {
	r := new(Replica)
	r.mu = new(sync.Mutex)
	r.log = make([]Entry, 0)
	r.peers = make([]*Clerk, len(peerHosts))
	r.leaderCond = sync.NewCond(r.mu)
	r.commitCond = sync.NewCond(r.mu)
	r.applyCond = sync.NewCond(r.mu)
	r.commitf = commitf

	r.mu.Lock() // XXX: this is here to make sure that we don't start processing any RPCs until after peers has been initialized
	r.StartServer(me)
	time.Sleep(time.Second * 3) // XXX: this is here just to give time for everything to start up before connecting to peers

	for i, peerHost := range peerHosts {
		r.peers[i] = MakeClerk(peerHost)
	}

	// FIXME: this is temporary
	r.isLeader = isLeader
	if r.isLeader {
		r.acceptedIndex = make([]uint64, len(peerHosts))
	}

	go r.applyThread()
	go r.commitThread()
	n := uint64(len(r.peers))
	for i := uint64(0); i < n; i++ {
		local_i := i
		go func() {
			r.replicaThread(local_i)
		}()
	}
	r.mu.Unlock()
	return r
}

func (r *Replica) StartServer(host uint64) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[TRY_APPEND] = func(rawReq []byte, rawRep *[]byte) {
		e := decodeUint64(rawReq)
		r.TryAppendRPC(e)
	}

	handlers[PREPARE] = func(rawReq []byte, rawRep *[]byte) {
		pn := decodeUint64(rawReq)
		rep := new(PrepareReply)
		r.PrepareRPC(pn, rep)
		*rawRep = encodePrepareReply(rep)
	}

	handlers[PROPOSE] = func(rawReq []byte, rawRep *[]byte) {
		args := decodeProposeArgs(rawReq)
		b := r.ProposeRPC(args.Pn, args.CommitIndex, args.Log)
		*rawRep = encodeBool(b)
	}
	s := rpc.MakeRPCServer(handlers)
	s.Serve(host, 1) // 1 == num workers
}
