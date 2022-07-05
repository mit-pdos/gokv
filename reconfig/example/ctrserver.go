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

type LogEntryExtra struct {
	completed *bool
	reply     *FetchAndAppendReply
	cond      *sync.Cond
}

type ValServer struct {
	s *pb.Server[LogEntryExtra]

	sid    uint64
	nextID uint64

	mu              *sync.Mutex
	appliedIndex    uint64
	val             []byte
	truncationLimit uint64
}

func (s *ValServer) applyThread() {
	var appliedIndex uint64 = 0
	for {
		// TODO: add no overflow assumption
		err, le := s.s.GetEntry(appliedIndex + 1)
		machine.Assert(err == pb.ENone)

		s.mu.Lock()
		if le.HaveExtra {
			// send the reply to the RPC goroutine
			e := le.Extra
			*e.completed = true
			e.reply.val = s.val
			e.reply.err = pb.ENone
			e.cond.Signal()
		}
		s.val = append(s.val, le.Op...)
		s.appliedIndex += 1

		// truncate everything that can be truncated
		if s.appliedIndex <= s.truncationLimit {
			s.mu.Unlock()
			s.s.Truncate(s.appliedIndex)
		} else {
			s.mu.Unlock()
		}

		s.mu.Unlock()
	}
}

// Returns an error and a value. If the error is ENone, then the  FetchAndAppend
// went through and the returned value is the value before appending.
func (cs *ValServer) FetchAndAppend(args []byte, reply *FetchAndAppendReply) {
	var op []byte = make([]byte, 0, 16)

	reply.err = pb.ENotPrimary // this is the error returned if the op doesn't get committed
	op = marshal.WriteBytes(op, args)

	var completed bool = false
	cond := sync.NewCond(cs.mu)

	err := cs.s.Propose(op,
		LogEntryExtra{
			completed: &completed,
			reply:     reply,
			cond:      cond,
		},
		func() { // cancel fn
			cs.mu.Lock()
			completed = true
			reply.err = pb.ENotPrimary
			cond.Signal()
			cs.mu.Unlock()
		},
	)

	if err != pb.ENone {
		reply.err = err
		return
	}

	// wait for operation to be completed (either committed, or removed from log)
	cs.mu.Lock()
	for !completed {
		cond.Wait()
	}
	cs.mu.Unlock()
}

func (cs *ValServer) getState() (uint64, []byte) {
	cs.mu.Lock()
	v := cs.val
	i := cs.appliedIndex
	cs.mu.Unlock()
	return i, v
}

func (cs *ValServer) setState(index uint64, val []byte) {
	cs.mu.Lock()
	cs.val = val
	cs.appliedIndex = index
	cs.mu.Unlock()
}

func StartValServer() {
	// cs := new(ValServer)
	// cs.ctr = 0
	// cs.s = rsm.MakeServer(cs.apply, cs.getState)
}
