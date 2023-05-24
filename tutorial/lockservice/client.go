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
	var l *Locked // XXX: need this because can't return from inside for loop in Goose

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
				l = &Locked{rpcCl: ck.rpcCl, id: id}
				break
			} else { // lockStatus == StatusNotGranted
				break
			}
		}
		if l != nil {
			break
		}
	}
	return l
}

func (l *Locked) Release() {
	for {
		if l.rpcCl.release(l.id) == 0 {
			break
		}
	}
}
