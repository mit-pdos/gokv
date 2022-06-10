package reconfig

import (
	"sync"
)

// This is the type of thing that can be replicated
type StateMachine struct {
	// Applies the given op to the current state, and returns the response for that
	// operation.
	// E.g. for a put, the state changes and the reply is nil.
	// For a get, the state is unchanged and the reply is the value of the desired key.
	Apply func([]byte) []byte

	// returns a marshalled snapshot of the current state
	GetState func() []byte

	// sets the state based on a marshalled snapshot
	SetState func([]byte)
}

type DurableLog struct {
	// Appends the given entry to the log and makes it durable before returning.
	Append func(entry []byte)

	// Allows for the log to be truncated up through (and including) the given index
	Truncate func(index uint64)

	InstallLog func(startIndex uint64, newLog []LogEntry)

	NextIndex func() uint64
}

type LogEntry = []byte

type Server struct {
	mu *sync.Mutex

	clerks    []*Clerk
	isPrimary bool

	// These two are the protocol state that is separate from the durable log
	cn     uint64
	sealed bool

	commitIndex uint64
	dlog DurableLog
}

// External function, called by users of this library.
// Tries to apply the given `op` to the state machine.
// If unable to apply the operation (e.g. if this server is not currently the
// primary), returns ENotPrimary. If successful, returns ENone.
func (s *Server) TryCommitOp(op OpType, reply *[]byte) Error {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return ENotPrimary
	}

	// now tell everyone else about the op
	args := &AppendArgs{cn: s.cn, entry: op, index: uint64(s.dlog.NextIndex())}

	for _, ck := range s.clerks {
		ck.DoOperation(args)
	}
	s.mu.Unlock()
	return ENone
}

// Internal RPC. Applies a single operation to a replica.
func (s *Server) AppendRPC(args *AppendArgs) Error {
	s.mu.Lock()
	if args.cn < s.cn { // ignore old requests
		s.mu.Unlock()
		return EStale
	} // else, we must have args.cn == s.cn

	if args.index < s.dlog.NextIndex() {
		s.mu.Unlock()
		return ENone // already accepted it
	} else if args.index > s.dlog.NextIndex() {
		s.mu.Unlock()
		return EAppendOutOfOrder
	}
	// else, must have args.index == s.startIndex + uint64(len(s.log))

	s.dlog.Append(args.entry)
	// TODO: trigger some sort of condition variable

	// make stuff durable
	s.mu.Unlock()
	return ENone
}

// Invoked by a failover controller to tell this server that it's the primary
// for configuration cn.
// Includes the (or, "a") final state of the previous config.
//
// XXX: the plan for config change is:
// a.) reserve a config number and the membership at the config number,
// 	   but do not activate it
// b.) seal the old config number
// c.) bring new config up to date with previous active config
// d.) set config number as active
// e.) tell primary it can start applying new operations
func (s *Server) BecomePrimary(args *BecomePrimaryArgs) Error {
	s.mu.Lock()
	if args.repArgs.cn <= s.cn { // ignore requests for old cn
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
		// The only way a replica will reject this is if it enters a
		// configuration higher than args.repArgs.cn. In that case, this primary
		// has been scooped.
		if ck.BecomeReplica(args.repArgs) != ENone {
			success = false
			break
			// s.mu.Unlock()
			// return ENotPrimary
		}
	}
	if !success {
		s.mu.Unlock()
		return ENotPrimary
	}

	// at this point, everyone is caught up.
	s.cn = args.repArgs.cn
	s.isPrimary = true
	s.mu.Unlock()
	return ENone
}

// Tell this server to become a backup in the new configuration cn, with the
// given state and osn
func (s *Server) BecomeReplica(args *BecomeReplicaArgs) Error {
	s.mu.Lock()
	if args.cn <= s.cn { // ignore requests for old cn
		s.mu.Unlock()
		return EStale
	}

	// s.sm.SetState(args.state)
	s.osn = args.osn
	s.cn = args.cn
	s.sealed = false

	return ENone
}

// Guarantees that a future GetState(cn) on any server will return state that
// is a superset of whatever will ever be committed in cn.
func (s *Server) Seal(cn uint64) Error {
	s.mu.Lock()
	if cn < s.cn { // already sealed for old cn's
		s.mu.Unlock()
		return ENone
	} else if cn == s.cn { // can seal, and be up-to-date
		s.sealed = true
		s.mu.Unlock()
		return ENone
	} else { // cn > s.cn
		// we could seal ourselves here, but then we wouldn't be up-to-date, so
		// future GetState()'s wouldn't return up-to-date stuff.
		// XXX: We can get rid of this error case by requiring that you only
		// call Seal() after bringing the server up-to-date in cn
		s.mu.Unlock()
		return EStale
	}
}

func (s *Server) GetState() *GetStateReply {
	s.mu.Lock()
	reply := new(GetStateReply)
	reply.cn = s.cn
	reply.state = s.sm.GetState()
	reply.osn = s.osn
	s.mu.Unlock()
	return reply
}

func MakeServer(apply func(OpType) []byte, getState func() []byte, setState func([]byte)) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.osn = 0
	s.cn = 0
	s.isPrimary = false
	s.sealed = false
	s.sm = &StateMachine{
		Apply:    apply,
		GetState: getState,
		SetState: setState,
	}

	return s
}
