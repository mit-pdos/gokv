package clerk

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/goose/machine"
	// "log"
)

type Clerk struct {
	confCk        *config.Clerk
	replicaClerks []*pb.Clerk
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

func (ck *Clerk) ApplyRo(op []byte) []byte {
	var ret []byte
	for {
		// pick a random server to read from
		j := machine.RandomUint64() % uint64(len(ck.replicaClerks))

		var err e.Error
		err, ret = ck.replicaClerks[j].ApplyRo(op)
		if err == e.None {
			break
		} else {
			// log.Println("Error during applyRo(): ", err)
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
