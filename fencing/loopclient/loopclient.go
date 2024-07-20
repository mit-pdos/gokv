package loopclient

import (
	"github.com/goose-lang/goose/machine"
	"github.com/mit-pdos/gokv/fencing/client"
	"github.com/mit-pdos/gokv/grove_ffi"
	"log"
)

func LoopOnKey(key uint64, config grove_ffi.Address) {

	ck := client.MakeClerk(config)

	var lowerBound uint64 = ck.FetchAndIncrement(key)
	for {
		v := ck.FetchAndIncrement(key)
		machine.Assert(v > lowerBound)
		if v%1000 == 0 {
			log.Printf("reached %d >= %d", key, v)
		}
		lowerBound = v
	}
}
