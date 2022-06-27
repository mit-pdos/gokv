package example

import (
	"sync"

	pb "github.com/mit-pdos/gokv/reconfig/replica"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
)

type FetchAndAppendReply struct {
	err pb.Error
	val []byte
}

type ValServer struct {
	s *pb.Server

	sid    uint64
	nextID uint64

	mu                 *sync.Mutex
	appliedIndex       uint64
	val                []byte
	appliedIndex_conds map[uint64]*sync.Cond // indexed by the index in the log

	replies map[uint64](*FetchAndAppendReply) // indexed by request ID from nextID
}

func (cs *ValServer) applyThread() {
	// this threads owns 1/2 of cs.appliedIndex, and cs.mu owns the other
	// half.
	for {
		// TODO: add no overflow assumption
		err, op := cs.s.GetEntry(cs.appliedIndex + 1)
		machine.Assert(err == pb.ENone)

		cs.mu.Lock()
		reply := cs.val
		cs.val = append(cs.val, op[16:]...) // first 16 bytes are opSID+opID

		// If a client request thread is waiting for this (opSID,opID), then
		// send them the reply.
		opSID, op2 := marshal.ReadInt(op)
		if opSID == cs.sid {
			opID, _ := marshal.ReadInt(op2)

			reply_ptr, ok := cs.replies[opID]
			if ok {
				reply_ptr.err = pb.ENone
				reply_ptr.val = reply
				delete(cs.replies, opID)
			} else {
				machine.Assert(false)
				// If the opSID matches, then there should definitely still be a
				// thread that's waiting for a reply. Putting this assert here
				// tests that.
			}
		}

		// If there's someone waiting for this index to get committed in the
		// log, wake them up.
		cv, ok := cs.appliedIndex_conds[cs.appliedIndex]
		if ok {
			delete(cs.appliedIndex_conds, cs.appliedIndex)
			cv.Signal()
		}

		cs.appliedIndex += 1
		cs.mu.Unlock()
		cs.s.Truncate(cs.appliedIndex)
	}
}

// Returns an error and a value. If the error is ENone, then the  FetchAndAppend
// went through and the returned value is the value before appending.
func (cs *ValServer) FetchAndAppend(args []byte, reply *FetchAndAppendReply) {
	var op []byte = make([]byte, 0, 16)

	reply.err = pb.ENotPrimary // this is the error returned if the op doesn't get committed

	// Need to set up a "callback" where the operation reply is put.
	// This needs to be done before invoking Propose(), because we won't hold
	// cs.mu.Lock() when we call Propose().
	cs.mu.Lock()
	op = marshal.WriteInt(op, cs.sid)

	opID := cs.nextID
	cs.nextID += 1

	op = marshal.WriteInt(op, opID)
	cs.replies[opID] = reply
	cs.mu.Unlock()
	op = marshal.WriteBytes(op, args)

	err, idx := cs.s.Propose(op)

	if err != pb.ENone {
		cs.mu.Lock()
		delete(cs.replies, opID)
		cs.mu.Unlock()
		// tell client to try elsewhere
		reply.err = err
		return
	}

	// wait for index to be committed
	cs.mu.Lock()
	if idx > cs.appliedIndex {
		cs.appliedIndex_conds[idx] = sync.NewCond(cs.mu)

		for idx > cs.appliedIndex {
			cs.appliedIndex_conds[idx].Wait()
		}
	}
	// At this point, either our operation has been signaled to us, or a
	// different op has been committed at that index, which means we're not the
	// primary anymore.

	cs.mu.Unlock()
}

// FIXME: return appliedIndex as well
func (cs *ValServer) getState() []byte {
	cs.mu.Lock()
	ret := cs.val
	cs.mu.Unlock()
	return ret
}

func (cs *ValServer) setState(enc_state []byte) {
	cs.mu.Lock()
	// ret := cs.val
	cs.mu.Unlock()
}

func StartValServer() {
	// cs := new(ValServer)
	// cs.ctr = 0
	// cs.s = rsm.MakeServer(cs.apply, cs.getState)
}
