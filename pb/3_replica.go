package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/goose/machine"
	"log"
	"sync"
)

// Ultimately, might make this be []byte
type LogEntry = byte

type ReplicaServer struct {
	mu   *sync.Mutex
	cn   uint64         // conf num
	conf *Configuration // XXX: this is currently unused, but later
	// should be used for reconnecting etc. at least by the primary
	isPrimary bool

	// confClerk     *ConfClerk
	replicaClerks []*ReplicaClerk

	opLog []LogEntry

	commitIdx  uint64
	commitCond *sync.Cond
	matchIdx   []uint64
}

// func (s *ReplicaServer) UpdateConfig() {
// v := s.confClerk.Get(0)
//
// // in principle, this could be updated concurrently with a different
// // UpdateConfig
// conf := DecodePBConfiguration(v.val)
// if s.me == conf.primary {
// s.cn = v.ver
// s.conf = DecodePBConfiguration(v.val)
// }
// }

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
		// log.Println("postAppendRPC")
		if s.matchIdx[i] < uint64(len(args.log)) {
			s.matchIdx[i] = uint64(len(args.log))

			// check if commitIdx can be increased
			m := min(s.matchIdx)
			if m > s.commitIdx {
				s.commitIdx = m
			}
		}
	}
	s.mu.Unlock()
}

// This should be invoked locally by services to attempt appending op to the
// log
func (s *ReplicaServer) StartAppend(op LogEntry) bool {
	s.mu.Lock()
	if !s.isPrimary {
		s.mu.Unlock()
		return false
	}
	s.opLog = append(s.opLog, op)
	s.matchIdx[0] = uint64(len(s.opLog)) // FIXME: trigger commitIdx update

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
		i := i
		go func() {
			ck.AppendRPC(args)
			s.postAppendRPC(uint64(i)+1, args)
		}()
	}
	return true
}

func (s *ReplicaServer) GetCommittedLog() []LogEntry {
	s.mu.Lock()
	// r := s.opLog
	r := s.opLog[:s.commitIdx]
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
		s.cn = args.cn
	}

	if args.commitIdx > s.commitIdx {
		s.commitIdx = args.commitIdx
	}
	s.isPrimary = false
	s.mu.Unlock()
	return true
}

// controller tells the primary to become the primary, and gives it the config
func (s *ReplicaServer) BecomePrimaryRPC(args *BecomePrimaryArgs) {
	s.mu.Lock()
	log.Printf("Becoming primary in %d, %+v", args.Cn, args.Conf.Replicas)

	if s.cn >= args.Cn {
		s.mu.Unlock()
		return
	}
	// FIXME(goose): should be able to put "else if" here
	if args.Cn > s.cn + 1 && s.cn == 0 {
		// XXX: if s.cn == 0 and e.g. I'm asked to become leader of
		// configuration 17, it's possible that stuff was committed in old
		// configurations, and I was added to the system but never got an
		// AppendRPC from a previous primary, so I have an incomplete log.
		// For now, just panic. Slightly better would be to return an error to
		// the controller.
		// Even better would be to have the configuration server *know* that
		// XXX: This assumes that a node can't be added, then removed, then
		// added again with its state intact. If this did happen, then the node
		// would have a positive s.cn (e.g. 13), see a BecomePrimary(cn=17) and
		// do it, even though it might have been kicked out in CN 13 and not
		// even know it. To allow for nodes to reenter with their state present,
		// we'd need the BecomePrimaryRPC to include a "StartCN" with the
		// property that the node has been in the system from args.StartCN up to
		// args.Cn. If the local number is at least StartCN, then the node can
		// successfully become primary, otherwise it can't.
		machine.Assume(false)
		return
	}
	s.isPrimary = true
	s.cn = args.Cn
	// s.conf = args.Conf
	s.matchIdx = make([]uint64, len(args.Conf.Replicas))

	s.replicaClerks = make([]*ReplicaClerk, len(args.Conf.Replicas) - 1)
	for i, _ := range args.Conf.Replicas[1:] {
		s.replicaClerks[i] = MakeReplicaClerk(args.Conf.Replicas[i+1])
	}
	s.mu.Unlock()
}

// used for recovery/adding a new node into the system
func (s *ReplicaServer) GetCommitLogRPC(_ []byte, reply *[]byte) {
	s.mu.Lock()
	*reply = s.opLog[:s.commitIdx]
	s.mu.Unlock()
}

func StartReplicaServer(me rpc.HostName) *ReplicaServer {
	// construct the ReplicaServer object
	s := new(ReplicaServer)
	s.mu = new(sync.Mutex)
	s.opLog = make([]LogEntry, 0)
	s.commitIdx = 0

	// s.confClerk = MakeConfClerk(confServer)
	// v := s.confClerk.Get(0)
	// s.cn = v.ver
	// s.conf = DecodePBConfiguration(v.val)
	s.cn = 0

	s.isPrimary = false

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
	handlers[REPLICA_BECOMEPRIMARY] = func(raw_args []byte, raw_reply *[]byte) {
		s.BecomePrimaryRPC(DecodeBecomePrimaryArgs(raw_args))
	}

	handlers[REPLICA_HEARTBEAT] = func(_ []byte, _ *[]byte) {
	}
	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)

	return s
}
