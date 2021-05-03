package main

import (
	"github.com/mit-pdos/gokv/memkv"
	"github.com/mit-pdos/gokv/dist_ffi"
	"flag"
	"log"
	"fmt"
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

	s := memkv.MakeMemKVCoordServer(dist_ffi.MakeAddress(host))
	me := dist_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	log.Printf("Started coordinator server on port %d; id %d", port, me)
	s.Start(me)
	select{}
}
