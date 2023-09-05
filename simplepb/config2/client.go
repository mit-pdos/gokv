package config2

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconnectclient"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
)

type Clerk struct {
	mu     *sync.Mutex
	cls    []*reconnectclient.ReconnectingClient
	leader uint64
}

const (
	RPC_RESERVEEPOCH   = uint64(0)
	RPC_GETCONFIG      = uint64(1)
	RPC_TRYWRITECONFIG = uint64(2)
	RPC_GETLEASE       = uint64(3)
)

func MakeClerk(hosts []grove_ffi.Address) *Clerk {
	var cls = make([]*reconnectclient.ReconnectingClient, 0)
	for _, host := range hosts {
		cls = append(cls, reconnectclient.MakeReconnectingClient(host))
	}
	return &Clerk{cls: cls, mu: new(sync.Mutex)}
}

func (ck *Clerk) ReserveEpochAndGetConfig() (uint64, []grove_ffi.Address) {
	reply := new([]byte)
	for {
		ck.mu.Lock()
		l := ck.leader
		ck.mu.Unlock()
		err := ck.cls[l].Call(RPC_RESERVEEPOCH, make([]byte, 0), reply, 100 /* ms */)
		if err != 0 {
			continue
		}

		var err2 uint64
		err2, *reply = marshal.ReadInt(*reply)
		if err2 == e.NotLeader {
			// potentially change leaders
			ck.mu.Lock()
			if l == ck.leader {
				ck.leader = (ck.leader + 1) % uint64(len(ck.cls))
			}
			ck.mu.Unlock()
			continue
		}
		if err2 == e.None {
			break
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
		i := machine.RandomUint64() % uint64(len(ck.cls))
		err := ck.cls[i].Call(RPC_GETCONFIG, make([]byte, 0), reply, 100 /* ms */)
		if err == 0 {
			break
		}
		continue
	}
	config := DecodeConfig(*reply)
	return config
}

func (ck *Clerk) TryWriteConfig(epoch uint64, config []grove_ffi.Address) e.Error {
	reply := new([]byte)
	var args = make([]byte, 0, 8+8*len(config))
	args = marshal.WriteInt(args, epoch)
	args = marshal.WriteBytes(args, EncodeConfig(config))
	// This has a high timeout because the server might need to wait for the
	// lease to expire before responding.

	for {
		ck.mu.Lock()
		l := ck.leader
		ck.mu.Unlock()

		err := ck.cls[l].Call(RPC_TRYWRITECONFIG, args, reply, 2000 /* ms */)
		if err != 0 {
			continue
		}
		err2, _ := marshal.ReadInt(*reply)

		if err2 == e.NotLeader {
			ck.mu.Lock()
			if l == ck.leader {
				ck.leader = (ck.leader + 1) % uint64(len(ck.cls))
			}
			ck.mu.Unlock()
			continue
		} else {
			break
		}
	}
	err, _ := marshal.ReadInt(*reply)
	return err
}

// returns e.None if the lease was granted for the given epoch, and a conservative
// guess on when the lease expires.
func (ck *Clerk) GetLease(epoch uint64) (e.Error, uint64) {
	reply := new([]byte)
	var args = make([]byte, 0, 8)
	args = marshal.WriteInt(args, epoch)

	for {
		ck.mu.Lock()
		l := ck.leader
		ck.mu.Unlock()

		err := ck.cls[l].Call(RPC_GETLEASE, args, reply, 100 /* ms */)
		if err != 0 {
			continue
		}
		err2, _ := marshal.ReadInt(*reply)

		if err2 == e.NotLeader {
			ck.mu.Lock()
			if l == ck.leader {
				ck.leader = (ck.leader + 1) % uint64(len(ck.cls))
			}
			ck.mu.Unlock()
			continue
		} else {
			break
		}
	}

	err2, enc := marshal.ReadInt(*reply)
	leaseExpiration, _ := marshal.ReadInt(enc)
	return err2, leaseExpiration
}
