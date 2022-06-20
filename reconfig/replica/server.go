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

	// Loads state from durable state
	Load func(name string) (uint64, uint64, uint64, []LogEntry, uint64)
}

type LogEntry = []byte

type Server struct {
	mu *sync.Mutex

	clerks    []*Clerk
	isPrimary bool

	// The in-memory state is here in place of "Getters".
	// We still need setters because we need to update durable state atomically
	// in various ways.

	epoch         uint64
	acceptedEpoch uint64
	startIndex    uint64
	log           []LogEntry

	dstate *DurableState

	// the commitIndex is not made durable (no need to)
	commitIndex uint64
}

// External function, called by users of this library.
// Tries to apply the given `op` to the state machine.
// If unable to apply the operation (e.g. if this server is not currently the
// primary), returns ENotPrimary. If successful, returns ENone.
// FIXME: also return index
func (s *Server) AppendOp(op LogEntry) Error {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return ENotPrimary
	}

	// now tell everyone else about the op
	args := &AppendArgs{epoch: s.epoch, entry: op, index: s.startIndex + uint64(len(s.log))}

	// TODO: pipelining; need to be able to send out a second AppendArgs RPC
	// even after the first one. Ideally, we only send the second one after the
	// first one has definitely been sent.

	// TODO: don't hold lock around this
	for _, ck := range s.clerks {
		ck.appendRPC(args)
	}
	s.mu.Unlock()
	return ENone
}

func (s *Server) isEpochStale(epoch uint64) bool {
	if epoch < s.epoch { // ignore old requests
		return true
	}
	return false
}

// Internal RPC. Applies a single operation to a replica.
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

	s.dstate.Append(args.epoch, args.entry)

	// make stuff durable
	s.mu.Unlock()
	return ENone
}

func (s *Server) EnterNewEpochRPC(epoch uint64) {
	s.mu.Lock()
	if s.epoch > epoch {
		s.mu.Unlock()
	}
	s.epoch = epoch
	s.mu.Unlock()
}

func (s *Server) BecomePrimary(args *BecomePrimaryArgs) Error {
	s.mu.Lock()
	s.isPrimary = true

	if s.isEpochStale(args.epoch) {
		s.mu.Unlock()
		return EStale
	}

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
		if ck.BecomeReplica(args.repArgs) != ENone {
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
func (s *Server) BecomeReplicaRPC(args *BecomeReplicaArgs) Error {
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

	s.epoch = args.epoch
	s.acceptedEpoch = args.epoch
	s.startIndex = args.startIndex
	s.log = args.log
	s.dstate.SetLog(args.epoch, args.startIndex, args.log)

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
