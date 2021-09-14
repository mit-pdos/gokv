package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/memkv"
	"log"
	"os"
)

func main() {
	var port uint64
	var host string
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.StringVar(&host, "init", "", "host for initial shard server")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := memkv.MakeKVCoordServer(grove_ffi.MakeAddress(host))
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	log.Printf("Started coordinator server on port %d; id %d", port, me)
	s.Start(me)
	select {}
}
