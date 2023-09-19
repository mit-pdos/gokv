package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config2"
	"log"
	"os"
)

func main() {
	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number to user for server; port + 1 is used for paxos")
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

	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	paxosMe := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port+1))
	config2.StartServer("config.data", me, paxosMe, []grove_ffi.Address{paxosMe}, servers)
	log.Printf("Started config server on port %d and %d; id %d", port, port+1, me)
	select {}
}
