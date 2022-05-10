package main

import (
	"flag"
	"github.com/mit-pdos/gokv/fencing/loopclient"
	"github.com/mit-pdos/gokv/grove_ffi"
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

	go loopclient.LoopOnKey(0, config)
	go loopclient.LoopOnKey(1, config)
	select {}
}
