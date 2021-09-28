package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
)

type LogEntry = []byte

type ReplicaServer struct {
	mu        *sync.Mutex
	opLog     []LogEntry
	commitIdx uint64

	cn        uint64 // conf num
	conf      *PBConfiguration
	isPrimary bool // optimization

	confClerk *ConfClerk
}

// This should be invoked locally by services to attempt appending op to the
// log
func (s *ReplicaServer) StartAppend(op LogEntry) bool {
	s.mu.Lock()
	if s.isPrimary {
		return false
		s.mu.Unlock()
	}
	s.opLog = append(s.opLog, op)
	s.mu.Unlock()
	return true
}

func (s *ReplicaServer) Append(op LogEntry) bool {
	s.mu.Lock()
	if s.isPrimary {
		return false
		s.mu.Unlock()
	}
	s.opLog = append(s.opLog, op)
	s.mu.Unlock()
	// FIXME: make AppendRPCs to all of the servers and collect their responses.
	// If any of them time out, kick them out of the system.
	return true
}

func (s *ReplicaServer) GetNextLogEntry() []byte {
	// FIXME: impl
	return nil
}

type AppendArgs struct {
	cn        uint64
	log       []LogEntry
	commitIdx uint64
}

func (s *ReplicaServer) AppendRPC(args AppendArgs) bool {
	s.mu.Lock()
	if s.cn != args.cn {
		// FIXME: if args.cn > s.conf.cn, then we should talk to the confserver
		s.mu.Unlock()
		return false
	}

	if uint64(len(args.log)) > uint64(len(s.opLog)) {
		s.opLog = args.log
	}
	if args.commitIdx > s.commitIdx {
		s.commitIdx = args.commitIdx
		// FIXME: signal commitIdx
	}
	s.mu.Unlock()
	return true
}

// used for recovery/adding a new node into the system
func (s *ReplicaServer) GetLogRPC(_ []byte, reply *[]byte) {
	s.mu.Lock()
	// FIXME: have to marshal this now...
	// *reply = s.opLog
	s.mu.Unlock()
}

func StartReplicaServer(me rpc.HostName, confServer rpc.HostName) *ReplicaServer {
	s := new(ReplicaServer)
	s.mu = new(sync.Mutex)
	s.opLog = make([]LogEntry, 0)
	s.commitIdx = 0
	s.confClerk = MakeConfClerk(confServer)
	v := s.confClerk.Get(0)
	s.cn = v.ver
	s.conf = DecodePBConfiguration(v.val)

	s.isPrimary = (me == s.conf.primary)
	return s
}
