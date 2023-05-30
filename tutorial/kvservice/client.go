package lockservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
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
		rpcCl: makeClient(host),
	}
}

func (ck *Clerk) Put(key string, val string) {
	var err uint64
	var opId uint64

	for {
		opId, err = ck.rpcCl.getFreshNumRpc()
		if err == 0 {
			break
		}
	}

	args := &putArgs{
		opId: opId,
		key:  key,
		val:  val,
	}

	for ck.rpcCl.putRpc(args) != urpc.ErrNone {
	}
}

// returns true if ConditionalPut was successful, false if current value did not
// match expected value.
func (ck *Clerk) ConditionalPut(key string, expectedVal string, newVal string) bool {
	var err uint64
	var opId uint64

	for {
		opId, err = ck.rpcCl.getFreshNumRpc()
		if err == 0 {
			break
		}
	}

	args := &conditionalPutArgs{
		opId:        opId,
		key:         key,
		expectedVal: expectedVal,
		newVal:      newVal,
	}

	var ret bool
	for {
		reply, err := ck.rpcCl.conditionalPutRpc(args)
		if err == urpc.ErrNone {
			ret = (reply == "ok")
			break
		}
	}
	return ret
}

// returns true if ConditionalPut was successful, false if current value did not
// match expected value.
func (ck *Clerk) Get(key string) string {
	var err uint64
	var opId uint64

	for {
		opId, err = ck.rpcCl.getFreshNumRpc()
		if err == 0 {
			break
		}
	}

	args := &getArgs{
		opId: opId,
		key:  key,
	}

	var ret string
	for {
		reply, err := ck.rpcCl.getRpc(args)
		if err == urpc.ErrNone {
			ret = reply
			break
		}
	}
	return ret
}
