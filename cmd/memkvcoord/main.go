package main

import (
	"github.com/mit-pdos/lockservice/grove_ffi"
	"github.com/mit-pdos/gokv/memkv"
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

	grove_ffi.SetPort(37000) // static port for memkv_coord
	log.Printf("Started coordinator server on port %d", port)
	s := memkv.MakeMemKVCoordServer()
	s.Start()
	select{}
}
