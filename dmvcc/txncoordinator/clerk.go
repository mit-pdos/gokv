package txncoordinator

// import "github.com/mit-pdos/gokv/grove_ffi"

type Clerk struct {
	s *Server
}

func (ck *Clerk) TryCommit(tid uint64, writes map[uint64]string) bool {
	return ck.s.TryCommit(tid, writes)
}

func MakeClerk(host *Server) *Clerk {
	return &Clerk{
		s: host,
	}
}
