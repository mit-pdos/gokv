package main

import (
	"github.com/mit-pdos/gokv/memkv"
	"fmt"
	"flag"
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

	log.Printf("Started coordinator server on port %d", port)
	s := memkv.MakeMemKVCoordServer("localhost:37001")
	s.Start(fmt.Sprintf("localhost:%d", port))
	select{}
}
