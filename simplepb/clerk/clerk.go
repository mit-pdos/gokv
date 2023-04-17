package clerk

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/mit-pdos/gokv/trusted_proph"
	"github.com/tchajed/goose/machine"
)

type Clerk struct {
	confCk           *config.Clerk
	replicaClerks    []*pb.Clerk
	preferredReplica uint64
}

func makeClerks(servers []grove_ffi.Address) []*pb.Clerk {
	clerks := make([]*pb.Clerk, len(servers))
	var i = uint64(0)
	for i < uint64(len(clerks)) {
		clerks[i] = pb.MakeClerk(servers[i])
		i += 1
	}
	return clerks
}

func Make(confHost grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.confCk = config.MakeClerk(confHost)
	for {
		config := ck.confCk.GetConfig()
		if len(config) == 0 {
			continue
		} else {
			ck.replicaClerks = makeClerks(config)
			break
		}
	}
	return ck
}

// will retry forever
func (ck *Clerk) Apply(op []byte) []byte {
	var ret []byte
	for {
		var err e.Error
		err, ret = ck.replicaClerks[0].Apply(op)
		if err == e.None {
			break
		} else {
			// log.Println("Error during apply(): ", err)
			machine.Sleep(uint64(100) * uint64(1_000_000)) // throttle retries to config server
			config := ck.confCk.GetConfig()
			if len(config) > 0 {
				ck.replicaClerks = makeClerks(config)
			}
			continue
		}
	}
	return ret
}

func (ck *Clerk) ApplyRo2(op []byte) []byte {
	var ret []byte
	for {
		// pick a random server to initially read from

		offset := ck.preferredReplica
		var err e.Error
		var i uint64
		// try all the servers starting from that random offset
		for i < uint64(len(ck.replicaClerks)) {
			k := (i + offset) % uint64(len(ck.replicaClerks))
			err, ret = ck.replicaClerks[k].ApplyRo(op)
			if err == e.None {
				ck.preferredReplica = k
				break
			}
			i += 1
		}

		if err == e.None {
			break
		} else {
			machine.Sleep(uint64(10) * uint64(1_000_000)) // throttle retries to config server
			config := ck.confCk.GetConfig()
			if len(config) > 0 {
				ck.replicaClerks = makeClerks(config)
				ck.preferredReplica = machine.RandomUint64() % uint64(len(ck.replicaClerks))
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
