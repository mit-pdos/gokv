package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
)

type LogEntry = []byte

type ReplicaServer struct {
	mu        *sync.Mutex
	cn        uint64 // conf num
	conf      *PBConfiguration
	isPrimary bool

	confClerk     *ConfClerk
	replicaClerks []*ReplicaClerk

	opLog []LogEntry

	commitIdx  uint64
	appliedIdx uint64
	commitCond *sync.Cond
	matchIdx []uint64
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
	s.mu.Unlock()
	return true
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

func (s *ReplicaServer) Append(op LogEntry) bool {
	s.mu.Lock()
	if s.isPrimary {
		s.mu.Unlock()
		return false
	}
	s.opLog = append(s.opLog, op)

	clerks := s.replicaClerks
	args := AppendArgs{
		cn:        s.cn,
		log:       s.opLog,
		commitIdx: s.commitIdx,
	}
	s.mu.Unlock()
	// FIXME: If any replicas time out, kick them out of the system.
	// TODO: replica can respond back to primary with newer config num

	for i, ck := range clerks {
		ck := ck // XXX: because goose doesn't support passing in parameters
		go func() {
			ck.AppendRPC(args)
			s.mu.Lock()
			if s.cn == args.cn {
				if s.matchIdx[i] < uint64(len(args.log)) {
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
		}()
	}
	return true
}

func (s *ReplicaServer) GetNextLogEntry() LogEntry {
	s.mu.Lock()
	for s.appliedIdx >= s.commitIdx {
		s.commitCond.Wait()
	}

	s.appliedIdx += 1
	ret := s.opLog[s.commitIdx]
	s.mu.Unlock()
	return ret
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
		s.commitCond.Signal()
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

func StartReplicaServer(me rpc.HostName, confServer rpc.HostName) {
	s := new(ReplicaServer)
	s.mu = new(sync.Mutex)
	s.opLog = make([]LogEntry, 0)
	s.commitIdx = 0
	s.confClerk = MakeConfClerk(confServer)
	v := s.confClerk.Get(0)
	s.cn = v.ver
	s.conf = DecodePBConfiguration(v.val)

	s.isPrimary = (me == s.conf.primary)
}
