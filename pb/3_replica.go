package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
)

// Ultimately, might make this be []byte
type LogEntry = byte

type ReplicaServer struct {
	mu        *sync.Mutex
	cn        uint64 // conf num
	conf      *PBConfiguration // XXX: this is currently unused, but later
	                           // should be used for reconnecting etc. at least
	                           // by the primary
	isPrimary bool

	confClerk     *ConfClerk
	replicaClerks []*ReplicaClerk

	opLog []LogEntry

	commitIdx  uint64
	commitCond *sync.Cond
	matchIdx   []uint64
}

func (s *ReplicaServer) UpdateConfig() {
	v := s.confClerk.Get(0)

	// in principle, this could be updated concurrently with a different
	// UpdateConfig
	if v.ver > s.cn {
		s.cn = v.ver
		s.conf = DecodePBConfiguration(v.val)
	}
}

func min(l []uint64) uint64 {
	var m uint64 = uint64(18446744073709551615)
	for _, v := range l {
		if v < m {
			m = v
		}
	}
	return m
}

func (s *ReplicaServer) postAppendRPC(i uint64, args *AppendArgs) {
	s.mu.Lock()
	if s.cn == args.cn {
		if s.matchIdx[i] > uint64(len(args.log)) {
			s.matchIdx[i] = uint64(len(args.log))

			// check if commitIdx can be increased
			m := min(s.matchIdx)
			if m > s.commitIdx {
				s.commitIdx = m
				s.commitCond.Signal()
			}
		}
	}
	s.mu.Unlock()
}

// This should be invoked locally by services to attempt appending op to the
// log
func (s *ReplicaServer) StartAppend(op LogEntry) bool {
	s.mu.Lock()
	if s.isPrimary {
		s.mu.Unlock()
		return false
	}
	s.opLog = append(s.opLog, op)

	clerks := s.replicaClerks
	args := &AppendArgs{
		cn:        s.cn,
		log:       s.opLog,
		commitIdx: s.commitIdx,
	}
	s.mu.Unlock()

	// XXX: use multipar?
	for i, ck := range clerks {
		ck := ck // XXX: because goose doesn't support passing in parameters
		go func() {
			ck.AppendRPC(args)
			s.postAppendRPC(uint64(i), args)
		}()
	}
	return true
}

func (s *ReplicaServer) GetCommittedLog() []LogEntry {
	s.mu.Lock()
	r := s.opLog[:s.commitIdx+1]
	s.mu.Unlock()
	return r
}

func (s *ReplicaServer) AppendRPC(args *AppendArgs) bool {
	s.mu.Lock()
	if s.cn > args.cn {
		s.mu.Unlock()
		return false
	}

	if s.cn < args.cn || uint64(len(args.log)) > uint64(len(s.opLog)) {
		s.opLog = args.log
	}
	s.cn = args.cn

	if args.commitIdx > s.commitIdx {
		s.commitIdx = args.commitIdx
	}
	s.mu.Unlock()
	return true
}

// controller tells the primary to become the primary, and gives it the config
func (s *ReplicaServer) BecomePrimary(args *BecomePrimaryArgs) {
	s.mu.Lock()
	if s.cn > args.cn {
		return
	}
	s.cn = args.cn
	s.conf = args.conf
	s.matchIdx = make([]uint64, len(args.conf.replicas))

	replicaClerks := make([]*ReplicaClerk, len(args.conf.replicas))
	for i, _ := range(s.conf.replicas) {
		replicaClerks[i] = MakeReplicaClerk(s.conf.replicas[i])
	}

	s.mu.Unlock()
}

// used for recovery/adding a new node into the system
func (s *ReplicaServer) GetCommitLogRPC(_ []byte, reply *[]byte) {
	s.mu.Lock()
	*reply = s.opLog[:s.commitIdx]
	s.mu.Unlock()
}

func StartReplicaServer(me rpc.HostName, confServer rpc.HostName) *ReplicaServer {
	// construct the ReplicaServer object
	s := new(ReplicaServer)
	s.mu = new(sync.Mutex)
	s.opLog = make([]LogEntry, 0)
	s.commitIdx = 0

	s.confClerk = MakeConfClerk(confServer)
	v := s.confClerk.Get(0)
	s.cn = v.ver
	s.conf = DecodePBConfiguration(v.val)

	s.isPrimary = (me == s.conf.primary)

	// Now start it
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[REPLICA_APPEND] = func(raw_args []byte, raw_reply *[]byte) {
		a := DecodeAppendArgs(raw_args)
		if s.AppendRPC(a) {
			*raw_reply = make([]byte, 1)
		} else {
			*raw_reply = make([]byte, 0)
		}
	}
	handlers[REPLICA_GETLOG] = s.GetCommitLogRPC

	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)

	return s
}
