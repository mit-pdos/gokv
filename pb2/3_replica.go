package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
)

// XXX: we're going to just make a new server object every time we want to
// change configurations.

type BackupServer struct {
	mu *sync.Mutex
	cn uint64 // conf num

	confClerk *ConfClerk

	opLog      []LogEntry
	commitIdx  uint64
	appliedIdx uint64
	commitCond *sync.Cond

	freshHost rpc.HostName // XXX: to be used when upgrading to a primary
}

func (s *BackupServer) GetNextLogEntry() LogEntry {
	s.mu.Lock()
	for s.appliedIdx >= s.commitIdx {
		s.commitCond.Wait()
	}

	s.appliedIdx += 1
	ret := s.opLog[s.commitIdx]
	s.mu.Unlock()
	return ret
}

func (s *BackupServer) AppendRPC(args AppendArgs) bool {
	s.mu.Lock()
	if s.cn != args.cn {
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

func StartBackupServer(me rpc.HostName, confServer rpc.HostName) {
	s := new(BackupServer)
	s.mu = new(sync.Mutex)
	s.opLog = make([]LogEntry, 0)
	s.commitIdx = 0
	s.confClerk = MakeConfClerk(confServer)
	v := s.confClerk.Get(0)
	s.cn = v.ver
}
