package lockservice

import (
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type Clerk struct {
	rpcCl *urpc.Client
}

type Locked struct {
	rpcCl      *urpc.Client
	locknum uint64
}

func (ck *Clerk) Acquire() *Locked {
	var reply []byte
	for {
		err := ck.rpcCl.Call(RPC_TRY_ACQUIRE, nil, &reply, 100 /* ms */)
		if err == 0 {
			break
		}
	}

	locknum, _ := marshal.ReadInt(reply)
	return &Locked{
		rpcCl: ck.rpcCl,
		locknum: locknum,
	}
}

func (l *Locked) Release() {
	//
}
