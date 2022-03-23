package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/fencing/ctr"
	"github.com/mit-pdos/gokv/grove_ffi"
)

func main() {
	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number of frontend server")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
		}
	}

	usage_assert(port != 0)

	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	ctr.StartServer(me)
	select{}
}
