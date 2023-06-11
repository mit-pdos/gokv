package kvservice

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

	for {
		// TODO: allocate these once, outside the loop. Waiting to do this until
		// heap has dfrac for convenience.
		args := &putArgs{
			opId: opId,
			key:  key,
			val:  val,
		}
		if ck.rpcCl.putRpc(args) == urpc.ErrNone {
			break
		}
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

	var ret bool
	for {
		args := &conditionalPutArgs{
			opId:        opId,
			key:         key,
			expectedVal: expectedVal,
			newVal:      newVal,
		}

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

	var ret string
	for {
		args := &getArgs{
			opId: opId,
			key:  key,
		}

		reply, err := ck.rpcCl.getRpc(args)
		if err == urpc.ErrNone {
			ret = reply
			break
		}
	}
	return ret
}
