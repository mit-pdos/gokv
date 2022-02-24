package reconfig

import (
	"sync"
)

type Server struct {
	mu        *sync.Mutex
	osn       uint64
	cn        uint64
	isPrimary bool
	clerks    []*Clerk
	sealed    bool

	// Applies the given op to the given state, and returns the response for that
	// operation.
	// E.g. for a put, the state changes and the reply is nil.
	// For a get, the state is unchanged and the reply is the value of the desired key.
	apply func(OpType)

	// returns a copy of the current state
	getState func() []byte

	// sets the state based on a marshalled snapshot
	setState func([]byte)
}

// External function, called by users of this library.
// Tries to apply the given `op` to the state machine.
// If unable to apply the operation (e.g. if this server is not currently the
// primary), returns ENotPrimary. If successful, returns ENone.
func (s *Server) PrimaryApplyOp(op OpType) Error {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return ENotPrimary
	}
	s.apply(op)
	s.osn += 1

	// now tell everyone else about the op
	args := &DoOperationArgs{cn: s.cn, op: op, osn: s.osn}
	for _, ck := range s.clerks {
		ck.DoOperation(args)
	}
	s.mu.Unlock()
	return ENone
}

// Internal RPC. Applies a single operation to a replica.
func (s *Server) DoOperation(args *DoOperationArgs) bool {
	s.mu.Lock()
	if args.cn < s.cn { // ignore old requests
		s.mu.Unlock()
		return false
	} // else, we must have args.cn == s.cn

	if args.osn <= s.osn {
		s.mu.Unlock()
		return true // already accepted it
	} // else, must have args.osn == s.osn+1

	s.apply(args.op)
	s.osn += 1

	s.mu.Unlock()
	return true
}

// Invoked by a failover controller to tell this server that it's the primary
// for configuration cn.
// Includes the (or, "a") final state of the previous config.
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

	s.setState(args.state)
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
	reply.state = s.getState()
	reply.osn = s.osn
	s.mu.Unlock()
	return reply
}

func MakeServer(apply func(OpType), getState func() []byte, setState func([]byte)) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.osn = 0
	s.cn = 0
	s.isPrimary = false
	s.sealed = false
	s.apply = apply
	s.getState = getState
	s.setState = setState

	return s
}
