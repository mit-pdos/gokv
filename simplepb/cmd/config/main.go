package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"log"
	"os"
)

func main() {
	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	a := flag.Args()
	servers := make([]grove_ffi.Address, 0)
	for _, srvStr := range a {
		servers = append(servers, grove_ffi.MakeAddress(srvStr))
	}

	s := config.MakeServer(servers)
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	s.Serve(me)
	log.Printf("Started config server on port %d; id %d", port, me)
	select {}
}
