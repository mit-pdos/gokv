package main

import (
	"flag"
	"github.com/mit-pdos/gokv/fencing/client"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/goose/machine"
)

func main() {
	var configStr string
	flag.StringVar(&configStr, "config", "", "address of config server")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
		}
	}

	usage_assert(configStr != "")

	config := grove_ffi.MakeAddress(configStr)

	ck := client.MakeClerk(config)

	x := func(key uint64) {
		var lowerBound uint64 = ck.FetchAndIncrement(key)
		for {
			v := ck.FetchAndIncrement(key)
			machine.Assert(v > lowerBound)
		}
	}

	go x(0)
	go x(1)
	select {}
}
