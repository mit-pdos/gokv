package main

import (
	"github.com/mit-pdos/gokv/memkv"
	"flag"
	"log"
	"os"
	"fmt"
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

	log.Printf("Started shard server on port %d", port)
	s := memkv.MakeMemKVShardServer()
	s.Start(fmt.Sprintf("localhost:%d", port))
	select{}
}
