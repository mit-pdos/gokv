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
	confCk    *config.Clerk
	primaryCk *pb.Clerk
}

func Make(confHost grove_ffi.Address) *Clerk {
	ck := new(Clerk)
	ck.confCk = config.MakeClerk(confHost)
	for {
		config := ck.confCk.GetConfig()
		if len(config) == 0 {
			continue
		} else {
			ck.primaryCk = pb.MakeClerk(config[0])
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
		err, ret = ck.primaryCk.Apply(op)
		if err == e.None {
			break
		} else {
			// log.Println("Error during apply(): ", err)
			machine.Sleep(uint64(100) * uint64(1_000_000)) // throttle retries to config server
			config := ck.confCk.GetConfig()
			if len(config) > 0 {
				ck.primaryCk = pb.MakeClerk(config[0])
			}
			continue
		}
	}
	return ret
}
