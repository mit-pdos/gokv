package index

// import "github.com/mit-pdos/gokv/grove_ffi"

type Clerk struct {
	s *Server
}

func (ck *Clerk) AcquireTuple(key uint64, tid uint64) uint64 {
	return ck.s.AcquireTuple(key, tid)
}

func (ck *Clerk) Read(key uint64, tid uint64) string {
	return ck.s.Read(key, tid)
}

func (ck *Clerk) UpdateAndRelease(tid uint64, writes map[uint64]string) {
	ck.s.UpdateAndRelease(tid, writes)
}

func MakeClerk(hostname *Server) *Clerk {
	return &Clerk{s: hostname}
}
