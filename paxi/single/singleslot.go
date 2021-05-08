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

func (r *Replica) TryDecide(v ValType) {
	pn := r.promisedPN + 1
	r.promisedPN = r.promisedPN + 1

	numPrepared := 0
	highestPn := uint64(0)
	var highestVal ValType
	highestVal = v // if no one in our majority has accepted a value, we'll propose this one
	mu := new(sync.Mutex)

	for _, peer := range r.peers {
		go func(peer *Clerk) {
			reply_ptr := new(PrepareReply)
			peer.Prepare(pn, reply_ptr) // TODO: replace with real RPC

			mu.Lock()
			if reply_ptr.Success {
				numPrepared = numPrepared + 1;
				if reply_ptr.Pn > highestPn {
					highestVal = reply_ptr.Val
					highestPn = reply_ptr.Pn
				}
			}
			mu.Unlock()
		}(peer)
	}

	mu.Lock()
	if numPrepared > len(r.peers)/2 {
		args := &ProposeArgs{Pn:pn, Val:highestVal}
		for _, peer := range r.peers {
			local_peer := peer // XXX: for Goose, I think
			go func() {
				r := local_peer.Propose(pn, highestVal) // TODO: replace with real RPC
			}()
		}
	} else {
		return
	}
}
