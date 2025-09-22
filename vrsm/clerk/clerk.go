package clerk

import (
	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/trusted_proph"
	"github.com/mit-pdos/gokv/vrsm/configservice"
	"github.com/mit-pdos/gokv/vrsm/replica"
	"github.com/mit-pdos/gokv/vrsm/replica/err_gk"
)

const (
	PreferenceRefreshTime = uint64(1_000_000_000) // 1 second
)

type Clerk struct {
	confCk                *configservice.Clerk
	replicaClerks         []*replica.Clerk
	preferredReplica      uint64
	lastPreferenceRefresh uint64
}

func makeClerks(servers []grove_ffi.Address) []*replica.Clerk {
	clerks := make([]*replica.Clerk, len(servers))
	var i = uint64(0)
	for i < uint64(len(clerks)) {
		clerks[i] = replica.MakeClerk(servers[i])
		i += 1
	}
	return clerks
}

func Make(confHosts []grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.confCk = configservice.MakeClerk(confHosts)
	for {
		config := ck.confCk.GetConfig()
		if len(config.Addrs) == 0 {
			continue
		} else {
			ck.replicaClerks = makeClerks(config.Addrs)
			break
		}
	}
	ck.preferredReplica = primitive.RandomUint64() % uint64(len(ck.replicaClerks))
	ck.lastPreferenceRefresh, _ = grove_ffi.GetTimeRange()
	return ck
}

// will retry forever
func (ck *Clerk) Apply(op []byte) []byte {
	var ret []byte
	for {
		var err err_gk.E
		err, ret = ck.replicaClerks[0].Apply(op)
		if err == err_gk.None {
			break
		} else {
			// log.Println("Error during apply(): ", err)
			primitive.Sleep(uint64(100) * uint64(1_000_000)) // throttle retries to config server
			config := ck.confCk.GetConfig()
			if len(config.Addrs) > 0 {
				ck.replicaClerks = makeClerks(config.Addrs)
			}
			continue
		}
	}
	return ret
}

func (ck *Clerk) maybeRefreshPreference() {
	now, _ := grove_ffi.GetTimeRange()
	if now > ck.lastPreferenceRefresh+PreferenceRefreshTime {
		ck.preferredReplica = primitive.RandomUint64() % uint64(len(ck.replicaClerks))
		ck.lastPreferenceRefresh, _ = grove_ffi.GetTimeRange()
	}
}

func (ck *Clerk) ApplyRo2(op []byte) []byte {
	var ret []byte
	ck.maybeRefreshPreference()
	for {
		// try to read initially from the "preferred" replica, then cycle around
		offset := ck.preferredReplica
		var err err_gk.E

		var i uint64
		// try all the servers starting from that random offset
		for i < uint64(len(ck.replicaClerks)) {
			k := (i + offset) % uint64(len(ck.replicaClerks))
			err, ret = ck.replicaClerks[k].ApplyRo(op)
			if err == err_gk.None {
				ck.preferredReplica = k
				break
			}
			i += 1
			ck.lastPreferenceRefresh, _ = grove_ffi.GetTimeRange()
			continue
		}

		if err == err_gk.None {
			break
		} else {
			timeToSleep := 5 + (primitive.RandomUint64() % 10)
			primitive.Sleep(timeToSleep * uint64(1_000_000)) // throttle retries to config server
			config := ck.confCk.GetConfig()
			if len(config.Addrs) > 0 {
				ck.replicaClerks = makeClerks(config.Addrs)
				ck.lastPreferenceRefresh, _ = grove_ffi.GetTimeRange()
				ck.preferredReplica = primitive.RandomUint64() % uint64(len(ck.replicaClerks))
			}
			continue
		}
	}
	return ret
}

func (ck *Clerk) ApplyRo(op []byte) []byte {
	p := trusted_proph.NewProph()
	v := ck.ApplyRo2(op)
	trusted_proph.ResolveBytes(p, v)
	return v
}
