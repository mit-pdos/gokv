package replica

import (
	"sync"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/reconfig/replica/appendargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/becomeprimaryargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/becomereplicaargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/error_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/getlogreply_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/logentry_gk"
)

type LogEntryAndExtra[ExtraT any] struct {
	Op        logentry_gk.S
	HaveExtra bool
	Extra     ExtraT

	haveCancel bool
	cancelFn   func()
}

type DurableState struct {
	// Append the entry to the end of the log
	Append func(entry logentry_gk.S)

	// Update the given three pieces of state atomically
	SetLog func(startIndex uint64, log []logentry_gk.S)

	// Allow for the prefix of the log up to and including the given `index` to be truncated
	Truncate func(index uint64)

	SetEpoch func(epoch uint64)
}

type ApplyFunc[ExtraT any] func(LogEntryAndExtra[ExtraT])

type Server[ExtraT any] struct {
	mu *sync.Mutex

	dstate *DurableState

	// This state matches the durable state in dstate
	epoch      uint64
	startIndex uint64
	log        []LogEntryAndExtra[ExtraT]

	// This state is not made durable
	isPrimary  bool
	clerks     []*Clerk
	matchIndex []uint64 // the primary remembers how much of the log all the replicas have accepted

	commitIndex      uint64     // the commitIndex is not made durable (no need to)
	commitIndex_cond *sync.Cond // signaled whenever commitIndex is updated
	applyFn          ApplyFunc[ExtraT]
}

func min(l []uint64) uint64 {
	m := ^uint64(0) // max uint64
	for _, x := range l {
		if x < m {
			m = x
		}
	}
	return m
}

func (s *Server[ExtraT]) applyThread() {
	var appliedIndex uint64 = 0
	for {
		// TODO: add no overflow assumption
		err, le := s.GetEntry(appliedIndex + 1)
		primitive.Assert(err == error_gk.ENone)
		s.applyFn(le)
	}
}

func (s *Server[ExtraT]) postSuccessfulAppendRPC(idx uint64, args *appendargs_gk.S) {
	s.mu.Lock()
	// Check if this node has moved on to a future epoch, in which case this
	// reply to AppendRPC should be ignored.
	if s.epoch != args.Epoch {
		s.mu.Unlock()
		return
	}

	// increase matchIndex
	if args.Index > s.matchIndex[idx] {
		s.matchIndex[idx] = args.Index
	}

	// TODO: can this min be lower than commitIndex across a config change?
	s.commitIndex = min(s.matchIndex)
	s.mu.Unlock()
}

// External function, called by users of this library.
// Tries to apply the given `op` to the state machine.
// If unable to apply the operation (e.g. if this server is not currently the
// primary), returns (ENotPrimary, 0).
// Otherwise, returns ENone and the index at which to expect the operation,
// along with the epoch at which it was proposed.
func (s *Server[ExtraT]) Propose(op logentry_gk.S, extra ExtraT, cancelFn func()) error_gk.E {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return error_gk.ENotPrimary
	}

	// now tell everyone else about the op
	index := s.startIndex + uint64(len(s.log))
	s.log = append(s.log, LogEntryAndExtra[ExtraT]{Op: op, Extra: extra, haveCancel: true, cancelFn: cancelFn})
	args := &appendargs_gk.S{Epoch: s.epoch, Entry: op, Index: index}
	clerks := s.clerks
	s.mu.Unlock()

	// TODO: pipelining; need to be able to send out a second AppendArgs RPC
	// even after the first one. Ideally, we only send the second one after the
	// first one has definitely been sent.
	for i, ck := range clerks {
		ck := ck
		idx := i
		go func() {
			// keep trying to get entry accepted until it succeeds, or until
			// we're no longer leader.
			for {
				err := ck.appendRPC(args)

				if err == error_gk.ENone {
					s.postSuccessfulAppendRPC(uint64(idx), args)
					break
				} else if err == error_gk.EStale {
					// we are no longer the leader in this epoch
					s.mu.Lock()
					if s.epoch == args.Epoch { // if we're still in the same epoch,
						s.isPrimary = false // stop telling people we're the leader
					}
					s.mu.Unlock()
					break
				} else if err == error_gk.EAppendOutOfOrder {
					// retry and hope that the missing append has managed to
					// reach the server now.
				}
				continue
			}
		}()
	}
	return error_gk.ENone
}

// Blocking function that waits for something to be committed at the given
// index, then returns what's been committed.
// Requires that the server's epoch number is at least `epoch`.
// Possible errors:
//
//	ETruncated iff the log has been truncated past the specified index.
//	EStale iff the server's epoch number is higher than the specified one.
func (s *Server[ExtraT]) GetEntry(index uint64) (error_gk.E, LogEntryAndExtra[ExtraT]) {
	s.mu.Lock()
	for s.commitIndex < index {
		s.commitIndex_cond.Wait()
	}

	if s.startIndex <= index {
		s.mu.Unlock()
		return error_gk.ENone, s.log[index-s.startIndex]
	} else {
		s.mu.Unlock()
		return error_gk.ETruncated, LogEntryAndExtra[ExtraT]{}
	}
}

func (s *Server[ExtraT]) Truncate(index uint64) {
	s.mu.Lock()
	if index >= uint64(len(s.log))+s.startIndex {
		s.log = make([]LogEntryAndExtra[ExtraT], 0)
	} else {
		s.log = s.log[index-s.startIndex:]
	}
	s.startIndex = index
	if s.commitIndex < s.startIndex {
		s.commitIndex = s.startIndex
	}
	s.dstate.Truncate(index)
	s.mu.Unlock()
}

// Internal RPC. Applies a single operation to a replica.
// Must have made sure that this replica has already entered the epoch.
func (s *Server[ExtraT]) appendRPC(args *appendargs_gk.S) error_gk.E {
	s.mu.Lock()

	if s.isEpochStale(args.Epoch) {
		s.mu.Unlock()
		return error_gk.EStale
	}
	// else, the epoch is up-to-date

	if args.Index < s.startIndex+uint64(len(s.log)) {
		s.mu.Unlock()
		return error_gk.ENone // already accepted it
	} else if args.Index > s.startIndex+uint64(len(s.log)) {
		s.mu.Unlock()
		return error_gk.EAppendOutOfOrder
	}
	// else, must have args.index == s.startIndex + uint64(len(s.log))

	s.log = append(s.log, LogEntryAndExtra[ExtraT]{Op: args.Entry})
	s.dstate.Append(args.Entry) // make stuff durable

	s.mu.Unlock()
	return error_gk.ENone
}

// Enters new epoch.
// returns true iff stale
func (s *Server[ExtraT]) isEpochStale(epoch uint64) bool {
	s.mu.Lock()
	if s.epoch > epoch {
		s.mu.Unlock()
		return true
	} else if s.epoch < epoch {
		s.commitIndex_cond.Broadcast()
		s.epoch = epoch
		s.dstate.SetEpoch(epoch)
		s.isPrimary = false
	}
	s.mu.Unlock()
	return false
}

// Must only be invoked after the primary has already entered the new epoch.
func (s *Server[ExtraT]) BecomePrimary(args *becomeprimaryargs_gk.S) error_gk.E {
	s.mu.Lock()
	if s.isEpochStale(args.Epoch) {
		s.mu.Unlock()
		return error_gk.EStale
	}
	if s.isPrimary {
		s.mu.Unlock() // Already primary, no work to do
		return error_gk.ENone
	}

	s.matchIndex = make([]uint64, len(args.Conf.Replicas))

	// make clerks
	s.clerks = make([]*Clerk, len(args.Conf.Replicas))
	for i := range s.clerks {
		s.clerks[i] = MakeClerk(args.Conf.Replicas[i])
	}

	s.isPrimary = true

	s.mu.Unlock()
	return error_gk.ENone
}

// TODO: put this in util file
func FmapList[T, S any](la []T, f func(T) S) []S {
	ret := make([]S, len(la))
	for i, a := range la {
		ret[i] = f(a)
	}
	return ret
}

// Tell this server to become a backup in the new configuration cn, with the
// given log and epoch.
//
// Might return EStale.
// Also, even if the epoch is not stale, the server might not have all the log
// entries it's supposed to keep around, in which case it returns
// EIncompleteLog and doesn't promise to have accepted args.log.
func (s *Server[ExtraT]) TryBecomeReplicaRPC(args *becomereplicaargs_gk.S) error_gk.E {
	s.mu.Lock()
	// if this is not a BRAND NEW epoch number, ignore it
	if args.Epoch <= s.epoch {
		s.mu.Unlock()
		return error_gk.EStale
	}
	s.epoch = args.Epoch

	s.isPrimary = false

	s.epoch = args.Epoch
	s.dstate.SetEpoch(args.Epoch)

	// XXX: We could do the following, but it's easier to only accept the log if
	// s.startIndex >= args.startIndex, since that should be the case most of
	// the time if doing a good reconfiguration.
	/*
		if args.startIndex > s.commitIndex {
			// server can't accept
			s.mu.Unlock()
			return EIncompleteLog
		}
	*/

	if args.StartIndex > s.startIndex {
		// server won't accept; technically it could if args.startIndex <=
		// s.commitIndex, see above comment.
		s.mu.Unlock()
		return error_gk.EIncompleteLog
	}

	prevLog := s.log
	s.log = FmapList(args.Log[s.startIndex-args.StartIndex:],
		func(e logentry_gk.S) LogEntryAndExtra[ExtraT] {
			return LogEntryAndExtra[ExtraT]{Op: e}
		})

	s.dstate.SetLog(args.StartIndex, args.Log)
	s.mu.Unlock()

	// conservatively cancel all old ops
	for _, le := range prevLog {
		if le.haveCancel {
			le.cancelFn()
		}
	}
	return error_gk.ENone
}

// Only possible error is EStale
func (s *Server[ExtraT]) RemainReplica(args *becomereplicaargs_gk.S) error_gk.E {
	s.mu.Lock()
	// if this is not a BRAND NEW epoch number, ignore it
	if args.Epoch <= s.epoch {
		s.mu.Unlock()
		return error_gk.EStale
	}
	s.epoch = args.Epoch
	s.dstate.SetEpoch(args.Epoch)

	s.isPrimary = false

	primitive.Assert(args.StartIndex < s.startIndex+uint64(len(s.log)))

	if args.StartIndex+uint64(len(args.Log)) < s.startIndex+uint64(len(s.log)) {
		// trim log
		prevLog := s.log
		s.log = s.log[:uint64(len(s.log))+args.StartIndex-s.startIndex]

		// cancel overwritten ops
		for _, le := range prevLog[uint64(len(s.log))+args.StartIndex-s.startIndex:] {
			if le.haveCancel {
				le.cancelFn()
			}
		}
	} else if args.StartIndex+uint64(len(args.Log)) > s.startIndex+uint64(len(s.log)) {
		// grow log
		s.log = append(s.log,
			FmapList(args.Log[s.startIndex+uint64(len(s.log))-args.StartIndex:], addDefaultExtra[ExtraT])...)
	}

	if args.StartIndex > s.commitIndex {
		s.commitIndex = args.StartIndex
		s.commitIndex_cond.Broadcast()
	}

	s.log = FmapList(args.Log, addDefaultExtra[ExtraT])

	s.dstate.SetLog(s.startIndex, FmapList(s.log, forgetExtra[ExtraT]))
	s.mu.Unlock()

	return error_gk.ENone
}

func forgetExtra[ExtraT any](l LogEntryAndExtra[ExtraT]) logentry_gk.S {
	return l.Op
}

func addDefaultExtra[ExtraT any](l logentry_gk.S) LogEntryAndExtra[ExtraT] {
	return LogEntryAndExtra[ExtraT]{Op: l}
}

func (s *Server[ExtraT]) GetUncommittedLog(epoch uint64) *getlogreply_gk.S {
	s.mu.Lock()
	reply := new(getlogreply_gk.S)
	if s.isEpochStale(epoch) {
		s.mu.Unlock()
		reply.Err = error_gk.EStale
		return reply
	}

	reply.Log = FmapList(s.log[(s.commitIndex-s.startIndex):], forgetExtra[ExtraT])
	reply.StartIndex = s.commitIndex
	reply.Err = error_gk.ENone

	s.mu.Unlock()
	return reply
}

type ProtocolState struct {
	epoch      uint64
	startIndex uint64
	log        []logentry_gk.S
}

func MakeServer[ExtraT any](dstate *DurableState, pstate *ProtocolState) *Server[ExtraT] {
	s := new(Server[ExtraT])
	s.mu = new(sync.Mutex)
	s.dstate = dstate

	s.epoch = pstate.epoch
	s.startIndex = pstate.startIndex
	s.log = FmapList(pstate.log,
		func(op logentry_gk.S) LogEntryAndExtra[ExtraT] {
			return LogEntryAndExtra[ExtraT]{Op: op}
		},
	)

	// We know that things are only truncated are being committed. Other than
	// that, we don't bother remembering anything about commitIndex
	s.commitIndex = s.startIndex

	// Even if we were the primary before crash+recovery, we'll forget about it
	// now and force a config change to happen.
	s.isPrimary = false

	return s
}
