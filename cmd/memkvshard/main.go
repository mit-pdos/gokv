package main

import (
	"github.com/mit-pdos/lockservice/grove_ffi"
	"github.com/mit-pdos/gokv/memkv"
	"flag"
	"log"
	"os"
)

func main() {
	var coord string
	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.StringVar(&coord, "coord", "", "address of coordinator")
	flag.Parse()

	if coord == "" || port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	grove_ffi.SetPort(port)
	log.Printf("Started shard server on port %d", port)
	s := memkv.MakeMemKVShardServer()
	s.Start()
	select{}
}
