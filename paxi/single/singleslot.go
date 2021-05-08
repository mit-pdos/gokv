package single

import (
	"sync"
)

// This isn't quite paxos
type Replica struct {
	promisedPN uint64 // server has promised not to accept proposals below this

	valPN uint64 // proposal number of accepted val
	acceptedVal ValType // the value itself

	committedVal ValType

	peers []*Clerk
}

type PrepareReply struct {
	Success bool
	Val uint64
	Pn uint64
}

func (r *Replica) PrepareRPC(pn uint64, reply *PrepareReply) {
	if pn > r.promisedPN {
		r.promisedPN = pn
		reply.Pn = r.valPN
		reply.Val = r.acceptedVal
		reply.Success = true
	}
	reply.Success = false
	reply.Pn = r.promisedPN
}

type ProposeArgs struct {
	Pn uint64
	Val ValType
}

func (r *Replica) ProposeRPC(args *ProposeArgs, reply *bool) {
	if args.Pn >= r.promisedPN {
		r.acceptedVal = args.Val
		r.valPN = args.Pn
		*reply = true
	} else {
		*reply = false
	}
}

// returns true iff there was an error
func (r *Replica) TryDecide(v ValType, outv *ValType) bool {
	pn := r.promisedPN + 1
	r.promisedPN = r.promisedPN + 1

	var numPrepared uint64
	numPrepared = 0
	var highestPn uint64
	highestPn = 0
	var highestVal ValType
	highestVal = v // if no one in our majority has accepted a value, we'll propose this one
	mu := new(sync.Mutex)

	for _, peer := range r.peers {
		local_peer := peer
		go func() {
			reply_ptr := new(PrepareReply)
			local_peer.Prepare(pn, reply_ptr) // TODO: replace with real RPC

			mu.Lock()
			if reply_ptr.Success {
				numPrepared = numPrepared + 1;
				if reply_ptr.Pn > highestPn {
					highestVal = reply_ptr.Val
					highestPn = reply_ptr.Pn
				}
			}
			mu.Unlock()
		}()
	}

	// FIXME: put this in a condvar loop with timeout
	mu.Lock()
	n := numPrepared
	mu.Unlock()

	if 2*n > uint64(len(r.peers)) {
		mu2 := new(sync.Mutex)
		var numAccepted uint64
		numAccepted = 0

		for _, peer := range r.peers {
			local_peer := peer // XXX: for Goose, I think
			// each thread talks to a unique peer
			go func() {
				r := local_peer.Propose(pn, highestVal) // TODO: replace with real RPC
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
			*outv = v
			return false
		} else {
			return true
		}
	} else {
		return true
	}
}
