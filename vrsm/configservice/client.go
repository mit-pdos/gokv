package configservice

import (
	"sync"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconnectclient"
	"github.com/mit-pdos/gokv/vrsm/configservice/config_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/err_gk"
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

func (ck *Clerk) ReserveEpochAndGetConfig() (uint64, config_gk.S) {
	reply := new([]byte)
	for {
		ck.mu.Lock()
		l := ck.leader
		ck.mu.Unlock()
		err := ck.cls[l].Call(RPC_RESERVEEPOCH, make([]byte, 0), reply, 100 /* ms */)
		if err != 0 {
			continue
		}

		var err2 err_gk.E
		err2, *reply = err_gk.Unmarshal(*reply)
		if err2 == err_gk.NotLeader {
			// potentially change leaders
			ck.mu.Lock()
			if l == ck.leader {
				ck.leader = (ck.leader + 1) % uint64(len(ck.cls))
			}
			ck.mu.Unlock()
			continue
		}
		if err2 == err_gk.None {
			break
		}
	}

	var epoch uint64
	epoch, *reply = marshal.ReadInt(*reply)
	config, _ := config_gk.Unmarshal(*reply)
	return epoch, config
}

func (ck *Clerk) GetConfig() config_gk.S {
	reply := new([]byte)
	for {
		i := primitive.RandomUint64() % uint64(len(ck.cls))
		err := ck.cls[i].Call(RPC_GETCONFIG, make([]byte, 0), reply, 100 /* ms */)
		if err == 0 {
			break
		}
		continue
	}
	config, _ := config_gk.Unmarshal(*reply)
	return config
}

func (ck *Clerk) TryWriteConfig(epoch uint64, config config_gk.S) err_gk.E {
	reply := new([]byte)
	var args = make([]byte, 0, 8+8*len(config.Addrs))
	args = marshal.WriteInt(args, epoch)
	args = config_gk.Marshal(args, config)
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
		err2, _ := err_gk.Unmarshal(*reply)

		if err2 == err_gk.NotLeader {
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
	err, _ := err_gk.Unmarshal(*reply)
	return err
}

// returns e.None if the lease was granted for the given epoch, and a conservative
// guess on when the lease expires.
func (ck *Clerk) GetLease(epoch uint64) (err_gk.E, uint64) {
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
		err2, _ := err_gk.Unmarshal(*reply)

		if err2 == err_gk.NotLeader {
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

	err2, enc := err_gk.Unmarshal(*reply)
	leaseExpiration, _ := marshal.ReadInt(enc)
	return err2, leaseExpiration
}
