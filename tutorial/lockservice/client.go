package lockservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/goose/machine"
)

type Clerk struct {
	rpcCl *Client
}

type Locked struct {
	rpcCl *Client
	id    uint64
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{
		rpcCl : makeClient(host),
	}
}

func (ck *Clerk) Acquire() *Locked {
	var id uint64
	var err uint64

	for {
		id, err = ck.rpcCl.getFreshNum()
		if err != 0 {
			continue
		}

		for {
			lockStatus, err := ck.rpcCl.tryAcquire(id)
			if err != 0 || lockStatus == StatusRetry {
				machine.Sleep(100 * 1_000_000) // 100 ms
				continue
			} else if lockStatus == StatusGranted {
				return &Locked{rpcCl: ck.rpcCl, id: id}
			} else { // lockStatus == StatusNotGranted
				break
			}
		}
	}
}

func (l *Locked) Release() {
	for {
		if l.rpcCl.release(l.id) == 0 {
			break
		}
	}
}
