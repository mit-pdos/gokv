package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/pb"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var port uint64

	// FIXME: confStr should be unnecessary, the config server can just talk to
	// the replica
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	// flag.StringVar(&coord, "coord", "", "address of coordinator")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	pb.StartConfServer(me)
	log.Printf("Started configuration service on port %d; id %d", port, me)
	select {}
}
