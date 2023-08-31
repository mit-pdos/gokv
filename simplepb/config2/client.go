package config2

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type Clerk struct {
	cl *urpc.Client // FIXME: use reconnectingclient
}

const (
	RPC_RESERVEEPOCH   = uint64(0)
	RPC_GETCONFIG      = uint64(1)
	RPC_TRYWRITECONFIG = uint64(2)
	RPC_GETLEASE       = uint64(3)
)

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) ReserveEpochAndGetConfig() (uint64, []grove_ffi.Address) {
	reply := new([]byte)
	for {
		// This has a high timeout because the server might need to wait for the
		// lease to expire before responding.
		err := ck.cl.Call(RPC_RESERVEEPOCH, make([]byte, 0), reply, 100 /* ms */)
		if err == 0 {
			break
		} else {
			continue
		}
	}

	var epoch uint64
	epoch, *reply = marshal.ReadInt(*reply)
	config := DecodeConfig(*reply)
	return epoch, config
}

func (ck *Clerk) GetConfig() []grove_ffi.Address {
	reply := new([]byte)
	for {
		err := ck.cl.Call(RPC_GETCONFIG, make([]byte, 0), reply, 100 /* ms */)
		if err == 0 {
			break
		} else {
			continue
		}
	}
	config := DecodeConfig(*reply)
	return config
}

func (ck *Clerk) TryWriteConfig(epoch uint64, config []grove_ffi.Address) e.Error {
	reply := new([]byte)
	var args = make([]byte, 0, 8+8*len(config))
	args = marshal.WriteInt(args, epoch)
	args = marshal.WriteBytes(args, EncodeConfig(config))
	err := ck.cl.Call(RPC_TRYWRITECONFIG, args, reply, 2000 /* ms */)
	if err == 0 {
		e, _ := marshal.ReadInt(*reply)
		return e
	} else {
		return err
	}
}

// returns e.None if the lease was granted for the given epoch, and a conservative
// guess on when the lease expires.
func (ck *Clerk) GetLease(epoch uint64) (e.Error, uint64) {
	reply := new([]byte)
	var args = make([]byte, 0, 8)
	args = marshal.WriteInt(args, epoch)
	err := ck.cl.Call(RPC_GETLEASE, args, reply, 100 /* ms */)
	if err == 0 {
		err2, enc := marshal.ReadInt(*reply)
		leaseExpiration, _ := marshal.ReadInt(enc)
		return err2, leaseExpiration
	} else {
		return err, 0
	}
}
