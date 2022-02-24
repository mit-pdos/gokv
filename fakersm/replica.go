package fakersm

import (
	"sync"
)

type LogEntry = []byte

type ReplicaServer struct {
	mu       *sync.Mutex
	firstIdx uint64
	log      []LogEntry
}

func (r *ReplicaServer) TryAppend(entry LogEntry) bool {
	r.mu.Lock()
	r.log = append(r.log, entry)
	r.mu.Unlock()
	return true
}

// XXX: precondition is that idx has not been truncated. Something funny going
// on here....
func (r *ReplicaServer) GetEntry(idx uint64) LogEntry {
	r.mu.Lock()
	// FIXME: if idx is too big, then do a CV wait.
	ret := r.log[idx]
	r.mu.Unlock()
	return ret
}

func (r *ReplicaServer) Recover() {
	r.mu.Lock()
	// no-op; this isn't fault tolerant!
	r.mu.Unlock()
}

func (r *ReplicaServer) Truncate(idx uint64) {
	r.mu.Lock()
	if idx > r.firstIdx {
		r.log = r.log[idx-r.firstIdx:]
	}
	r.mu.Unlock()
}

func MakeReplicaServer() *ReplicaServer {
	return &ReplicaServer{
		mu:       new(sync.Mutex),
		firstIdx: uint64(0),
		log:      make([]LogEntry, 0),
	}
}
