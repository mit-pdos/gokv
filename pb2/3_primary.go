package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
)

type LogEntry = []byte

type PrimaryServer struct {
	mu *sync.Mutex
	cn uint64 // conf num

	confClerk *ConfClerk

	opLog      []LogEntry
	commitIdx  uint64
	commitCond *sync.Cond
	appliedIdx uint64

	hasBackup  bool
	backup     *BackupClerk
	backupIdx   uint64
}

// This should be invoked locally by services to attempt appending op to the
// log
func (s *PrimaryServer) StartAppend(op LogEntry) bool {
	s.mu.Lock()
	s.opLog = append(s.opLog, op)
	s.mu.Unlock()
	return true
}

func (s *PrimaryServer) Append(op LogEntry) bool {
	s.mu.Lock()
	s.opLog = append(s.opLog, op)

	if !s.hasBackup {
		s.commitIdx = uint64(len(s.opLog))
		s.commitCond.Signal()
		s.mu.Unlock()
		return true
	}

	args := AppendArgs{
		cn:        s.cn,
		log:       s.opLog,
		commitIdx: s.commitIdx,
	}
	backup := s.backup
	s.mu.Unlock()

	backup.AppendRPC(args)
	s.mu.Lock()
	if s.cn == args.cn {
		if s.backupIdx < uint64(len(args.log)) {
			s.backupIdx = uint64(len(args.log))
			if s.backupIdx > s.commitIdx {
				s.commitIdx = s.backupIdx
				s.commitCond.Signal()
			}
		}
	}
	return true
}

func (s *PrimaryServer) GetNextLogEntry() LogEntry {
	s.mu.Lock()
	for s.appliedIdx >= s.commitIdx {
		s.commitCond.Wait()
	}

	s.appliedIdx += 1
	ret := s.opLog[s.commitIdx]
	s.mu.Unlock()
	return ret
}

// used for recovery/adding a new node into the system
func (s *PrimaryServer) GetLogRPC(_ []byte, reply *[]byte) {
	s.mu.Lock()
	// FIXME: have to marshal this now...
	// *reply = s.opLog
	s.mu.Unlock()
}

func StartPrimaryServer(me rpc.HostName, confServer rpc.HostName) {
	s := new(PrimaryServer)
	s.mu = new(sync.Mutex)
	s.opLog = make([]LogEntry, 0)
	s.commitIdx = 0
	s.confClerk = MakeConfClerk(confServer)
	v := s.confClerk.Get(0)
	s.cn = v.ver
}
