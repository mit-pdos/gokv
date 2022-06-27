package reconfig

import (
	"sync"
)

type DurableState struct {
	// Atomically append the entry to the end of the log and set acceptedEpoch
	// to the given epoch.
	Append func(acceptedEpoch uint64, entry []byte)

	// Update the given three pieces of state atomically
	SetLog func(acceptedEpoch uint64, startIndex uint64, log []LogEntry)

	// Allow for the prefix of the log up to and including the given `index` to be truncated
	Truncate func(index uint64)

	SetEpoch func(epoch uint64)
}

type LogEntry = []byte

type Server struct {
	mu *sync.Mutex

	dstate *DurableState

	// This state matches the durable state in dstate
	epoch         uint64
	acceptedEpoch uint64
	startIndex    uint64
	log           []LogEntry

	// This state is not made durable
	isPrimary bool
	clerks    []*Clerk

	commitIndex      uint64     // the commitIndex is not made durable (no need to)
	commitIndex_cond *sync.Cond // signaled whenever commitIndex is updated
	matchIndex       []uint64   // the primary remembers how much of the log all the replicas have accepted
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

func (s *Server) postSuccessfulAppendRPC(idx uint64, args *AppendArgs) {
	s.mu.Lock()
	// Check if this node has moved on to a future epoch, in which case this
	// reply to AppendRPC should be ignored.
	if s.epoch != args.epoch {
		s.mu.Unlock()
		return
	}

	// increase matchIndex
	if args.index > s.matchIndex[idx] {
		s.matchIndex[idx] = args.index
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
//
// TODO: do we want to give this epoch back out? An alternative is to insist
// that the client figure out how to give operations unique IDs. But, we have a
// unique identifier handy, so why not expose it?
func (s *Server) Propose(op LogEntry) (Error, uint64) {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return ENotPrimary, 0
	}

	// now tell everyone else about the op
	idx := s.startIndex + uint64(len(s.log))
	args := &AppendArgs{epoch: s.epoch, entry: op, index: idx}

	// TODO: pipelining; need to be able to send out a second AppendArgs RPC
	// even after the first one. Ideally, we only send the second one after the
	// first one has definitely been sent.

	clerks := s.clerks
	s.mu.Unlock()

	for i, ck := range clerks {
		ck := ck
		idx := i
		go func() {
			// keep trying to get entry accepted until it succeeds, or until
			// we're no longer leader.
			for {
				err := ck.appendRPC(args)

				if err == ENone {
					s.postSuccessfulAppendRPC(uint64(idx), args)
					break
				} else if err == EStale {
					// we are no longer the leader in this epoch
					s.mu.Lock()
					if s.epoch == args.epoch { // if we're still in the same epoch,
						s.isPrimary = false // stop telling people we're the leader
					}
					s.mu.Unlock()
					break
				} else if err == EAppendOutOfOrder {
					// retry and hope that the missing append has managed to
					// reach the server now.
				}
				continue
			}
		}()
	}

	return ENone, idx
}

// Blocking function that waits for something to be committed at the given
// index, then returns what's been committed.
// Possible errors:
//   ETruncated iff the log has been truncated past the specified index.
//
// TODO: (related to above todo) this could return the epoch number at which the
// log entry was proposed for deduplication. But that would mean that each entry
// would keep an epoch number, which is an extra 8 bytes on disk for each entry.
func (s *Server) GetEntry(index uint64) (Error, LogEntry) {
	s.mu.Lock()
	for s.commitIndex < index {
		s.commitIndex_cond.Wait()
	}

	if s.startIndex <= index {
		s.mu.Unlock()
		return ENone, s.log[index-s.startIndex]
	} else {
		s.mu.Unlock()
		return ETruncated, make([]byte, 0)
	}
}

func (s *Server) Truncate(index uint64) {
	s.mu.Lock()
	if index >= uint64(len(s.log))+s.startIndex {
		s.log = make([]LogEntry, 0)
	} else {
		s.log = s.log[index-s.startIndex:]
	}
	s.startIndex = index

	s.dstate.Truncate(index)
	s.mu.Unlock()
}

func (s *Server) isEpochStale(epoch uint64) bool {
	if epoch < s.epoch { // ignore old requests
		return true
	}
	return false
}

// Internal RPC. Applies a single operation to a replica.
// Must have made sure that this replica has already entered the epoch.
func (s *Server) appendRPC(args *AppendArgs) Error {
	s.mu.Lock()

	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return EStale
	}
	// else, the epoch is up-to-date

	if args.index < s.startIndex+uint64(len(s.log)) {
		s.mu.Unlock()
		return ENone // already accepted it
	} else if args.index > s.startIndex+uint64(len(s.log)) {
		s.mu.Unlock()
		return EAppendOutOfOrder
	}
	// else, must have args.index == s.startIndex + uint64(len(s.log))

	s.log = append(s.log, args.entry)
	s.dstate.Append(args.epoch, args.entry) // make stuff durable

	s.mu.Unlock()
	return ENone
}

func (s *Server) EnterNewEpochRPC(epoch uint64) {
	s.mu.Lock()
	if s.epoch > epoch {
		s.mu.Unlock()
	}
	s.epoch = epoch
	s.dstate.SetEpoch(epoch)
	s.isPrimary = false
	s.mu.Unlock()
}

// Must only be invoked after the primary has already entered the new epoch.
func (s *Server) BecomePrimary(args *BecomePrimaryArgs) Error {
	s.mu.Lock()
	if s.isPrimary {
		s.mu.Unlock() // Already
		return ENone
	}
	s.isPrimary = true

	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return EStale
	}

	s.matchIndex = make([]uint64, len(args.conf.replicas))

	// make clerks
	s.clerks = make([]*Clerk, len(args.conf.replicas))
	for i := range s.clerks {
		s.clerks[i] = MakeClerk(args.conf.replicas[i])
	}

	// tell the replicas to become replicas
	var success = true
	for _, ck := range s.clerks {
		// The only way a replica will reject this is if it knows about a higher
		// epoch number. In that case, this primary has been scooped.
		if ck.BecomeReplicaRPC(args.repArgs) != ENone {
			success = false
			break
		}
	}
	if !success {
		s.mu.Unlock()
		return ENotPrimary
	}

	// at this point, everyone is caught up.
	s.mu.Unlock()
	return ENone
}

// Tell this server to become a backup in the new configuration cn, with the
// given log and epoch.
//
// The server might be unable to become a replica server.
func (s *Server) TryBecomeReplicaRPC(args *BecomeReplicaArgs) Error {
	s.mu.Lock()
	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return EStale
	}

	if args.epoch <= s.acceptedEpoch {
		s.mu.Unlock()
		return ENone // XXX: this can return ENone, because if
		// s.acceptedEpoch = args.epoch, then this server will already have
		// accepted a "safe" proposal from the leader, and that's all that the
		// leader needs to know. I.e. the post-condition of BecomeReplicaRPC is
		// that this replica has accepted some safe proposal from the given
		// epoch.
	}

	s.isPrimary = false

	s.epoch = args.epoch
	s.dstate.SetEpoch(args.epoch)

	s.acceptedEpoch = args.epoch
	s.startIndex = args.startIndex
	s.log = args.log
	s.dstate.SetLog(args.epoch, args.startIndex, args.log)
	s.mu.Unlock()

	return ENone
}

func (s *Server) GetUncommittedLog() *GetLogReply {
	s.mu.Lock()
	reply := new(GetLogReply)

	reply.epoch = s.acceptedEpoch

	reply.log = s.log[(s.commitIndex - s.startIndex):]
	reply.startIndex = s.commitIndex

	s.mu.Unlock()
	return reply
}

type ProtocolState struct {
	epoch         uint64
	acceptedEpoch uint64
	startIndex    uint64
	log           []LogEntry
}

func MakeServer(dstate *DurableState, pstate *ProtocolState) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.dstate = dstate

	s.epoch = pstate.epoch
	s.acceptedEpoch = pstate.acceptedEpoch
	s.startIndex = pstate.startIndex
	s.log = pstate.log

	// We know that things are only truncated are being committed. Other than
	// that, we don't bother remembering anything about commitIndex
	s.commitIndex = s.startIndex

	// Even if we were the primary before crash+recovery, we'll forget about it
	// now and force a config change to happen.
	s.isPrimary = false

	return s
}
