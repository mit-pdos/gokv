package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/state"
	"log"
	"os"
)

func main() {
	var port uint64
	var fname string
	flag.Uint64Var(&port, "port", 0, "port number to user for server; (port-1) is used for pb server")
	flag.StringVar(&fname, "fname", "", "filename to put durable state in/to recover from")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := state.MakeServer(fname)
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	pbHost := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port-1))
	s.Serve(me, pbHost)
	log.Printf("Started kvserver on port %d; id %d", port, me)
	log.Printf("Started pbserver on port %d; id %d", port-1, pbHost)
	select {}
}
