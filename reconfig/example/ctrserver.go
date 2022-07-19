package example

import (
	"sync"

	pb "github.com/mit-pdos/gokv/reconfig/replica"
	// "github.com/tchajed/goose/machine"
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
	s.mu.Lock()
	for {
		// TODO: add no overflow assumption
		appliedIndex := s.appliedIndex
		s.mu.Unlock()

		err, le := s.s.GetEntry(appliedIndex + 1)
		if err == pb.ETruncated { // this is supposed to only happen when a snapshot is installed
			continue
		}

		s.mu.Lock()
		if s.appliedIndex != appliedIndex {
			// this means entries have been skipped
			continue
		}

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
		}
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

func (cs *ValServer) truncateAndBecomeReplica(args *pb.BecomeReplicaArgs) pb.Error {
	// FIXME: what if this request is from an old reconf attempt that failed?
	// If this replica is in the latest config, then this truncation might trim
	// off entries that this node has not yet applied, resulting in this node
	// not having the latest state, a violation of the informal invariant we
	// hope to maintain about the latest config.
	cs.mu.Lock()
	cs.appliedIndex = args.StartIndex
	cs.mu.Unlock()
	cs.s.Truncate(args.StartIndex)

	return cs.s.TryBecomeReplicaRPC(args)
}

func NewValServer() *ValServer {
	s := new(ValServer)
	// FIXME: init
	return s
}
