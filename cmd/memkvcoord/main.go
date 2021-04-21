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
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := memkv.MakeMemKVCoordServer(dist_ffi.MakeAddress("127.0.0.1:37002"))
	me := dist_ffi.MakeAddress(fmt.Sprintf("127.0.0.1:%d", port))
	log.Printf("Started coordinator server on port %d; id %d", port, me)
	s.Start(me)
	select{}
}
