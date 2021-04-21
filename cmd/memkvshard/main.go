package main

import (
	"github.com/mit-pdos/gokv/memkv"
	"github.com/mit-pdos/gokv/dist_ffi"
	"fmt"
	"flag"
	"log"
	"os"
)

func main() {
	// var coord string
	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	// flag.StringVar(&coord, "coord", "", "address of coordinator")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := memkv.MakeMemKVShardServer()
	me := dist_ffi.MakeAddress(fmt.Sprintf("127.0.0.1:%d", port))
	log.Printf("Started shard server on port %d; id %d", port, me)
	s.Start(me)
	select{}
}
