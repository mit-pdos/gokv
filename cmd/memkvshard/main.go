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
	var is_init bool
	var port uint64
	flag.BoolVar(&is_init, "init", false, "true iff this server owns all shard at initialization; default is false")
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	// flag.StringVar(&coord, "coord", "", "address of coordinator")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := memkv.MakeMemKVShardServer(is_init)
	me := dist_ffi.MakeAddress(fmt.Sprintf("127.0.0.1:%d", port))
	log.Printf("Started shard server on port %d; id %d", port, me)
	s.Start(me)
	select{}
}
