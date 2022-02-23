package reconfig

import (
	"sync"
)

type Server struct {
	mu        *sync.Mutex
	state     *[]byte
	osn       uint64
	cn        uint64
	isPrimary bool
	clerks    []*Clerk
}

type OpType = []byte

type Error = uint64

const (
	ENone       = uint64(0)
	ENotPrimary = uint64(1)
)

func ApplyOp(state *[]byte, op OpType) {
	// no-op for now; this ought to be determined by the user
}

// Tries to apply the given `op` to the state machine.
// If unable to apply the operation (e.g. if this server is not currently the
// primary), returns ENotPrimary. If successful, returns ENone.
func (s *Server) PrimaryApplyOp(op OpType) Error {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return ENotPrimary
	}
	ApplyOp(s.state, op)
	s.osn += 1

	// now tell everyone else about the op
	args := &DoOperationArgs{cn:s.cn, op:op, osn:s.osn}
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

	ApplyOp(s.state, args.op)
	s.osn += 1

	s.mu.Unlock()
	return true
}

// Invoked by a failover controller to tell this server that it's the primary
// for configuration cn.
// Includes the (or, "a") final state of the previous config.
func (s *Server) BecomePrimary(args *BecomePrimaryArgs) {
	s.mu.Lock()
	if args.repArgs.cn <= s.cn { // ignore requests for old cn
		s.mu.Unlock()
		return
	}

	// make clerks
	s.clerks = make([]*Clerk, len(args.conf.replicas))
	for i, _ := range s.clerks {
		s.clerks[i] = MakeClerk(args.conf.replicas[i])
	}

	s.cn = args.repArgs.cn
	s.mu.Unlock()
}

// Tell the server to become a backup in the new configuration cn, with the given state and osn
func (s *Server) BecomeReplica(args *BecomeReplicaArgs) {
}

func (s *Server) Seal(cn uint64) {
}

func (s *Server) GetState(cn uint64, state []byte, osn []byte) {
}
