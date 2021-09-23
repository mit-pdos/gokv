package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"sync"
)

type PBConfiguration struct {
	cn       uint64
	replicas []rpc.HostName
	primary  rpc.HostName
}

type ReplicaServer struct {
	mu        *sync.Mutex
	opLog     []byte
	commitIdx uint64
	conf      *PBConfiguration
	isPrimary bool // optimization
}

// This should be invoked locally by services to attempt appending op to the
// log
func (s *ReplicaServer) Append(op []byte) bool {
	s.mu.Lock()
	if s.isPrimary {
		return false
		s.mu.Unlock()
	}
	s.opLog = append(s.opLog, op...)
	s.mu.Unlock()
	return true
}

type AppendArgs struct {
	cn        uint64
	logid     uint64
	entries   []byte
	commitIdx uint64
}

func (s *ReplicaServer) AppendRPC(args AppendArgs) {
	s.mu.Lock()
	if s.conf.cn != args.cn {
		// FIXME: if args.cn > s.conf.cn, then we should talk to the confserver
		s.mu.Unlock()
		return false
	}

	if args.logid+uint64(len(args.entries)) > len(s.opLog) {
		// XXX: can prove that len(s.opLog) >= logid(!)
		s.opLog = append(s.opLog, args.entries[len(s.opLog)-args.logid:])
	}
	if args.commitIdx > s.commitIdx {
		s.commitIdx = args.commitIdx
	}
	s.mu.Unlock()
	return true
}

func (s *ReplicaServer) CatchUpRPC(log []byte) {
	s.mu.Lock()
	// FIXME: impl
	s.mu.Unlock()
}
