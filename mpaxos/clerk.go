package mpaxos

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/goose/machine"
)

type Clerk struct {
	cks []*singleClerk
}

func MakeClerk(config []grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.cks = make([]*singleClerk, len(config))
	var i uint64 = 0
	n := uint64(len(config))
	for i < n {
		ck.cks[i] = makeSingleClerk(config[i])
		i += 1
	}

	return ck
}

// XXX: not trying to prevent livelock right now. A client op simply starts by
// telling the server to become leader (which it might already be), then it
// tells to apply the op. Concurrent clients trying will easily result in
// livelock, where two nodes keep trying to become leader, but can't stay leader
// for long enough to process even one request.
func (ck *Clerk) Apply(op []byte) []byte {
	var reply []byte
	var err Error
	for {
		i := machine.RandomUint64() % uint64(len(ck.cks))
		cl := ck.cks[i]
		cl.becomeLeader()

		err, reply = cl.apply(op)
		if err == ENone {
			break
		} else {
			continue
		}
	}
	return reply
}
