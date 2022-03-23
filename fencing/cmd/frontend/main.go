package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/fencing/frontend"
	"github.com/mit-pdos/gokv/grove_ffi"
)

func main() {
	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number of frontend server")

	var configStr string
	flag.StringVar(&configStr, "config", "", "address of config server")

	var ctr1Str string
	flag.StringVar(&ctr1Str, "ctr1", "", "address of counter server 1")

	var ctr2Str string
	flag.StringVar(&ctr2Str, "ctr2", "", "address of counter server 2")

	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
		}
	}

	usage_assert(configStr != "")
	usage_assert(port != 0)
	usage_assert(ctr1Str != "")
	usage_assert(ctr2Str != "")

	config := grove_ffi.MakeAddress(configStr)
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	ctr1 := grove_ffi.MakeAddress(ctr1Str)
	ctr2 := grove_ffi.MakeAddress(ctr2Str)
	frontend.StartServer(me, config, ctr1, ctr2)
	select {}
}
