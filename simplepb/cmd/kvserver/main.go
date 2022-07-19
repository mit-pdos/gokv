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
	flag.Uint64Var(&port, "port", 0, "port number to use for server")
	flag.StringVar(&fname, "fname", "", "filename to put durable state in/to recover from")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := state.MakeServer(fname)
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	s.Serve(me)
	log.Printf("Started kvserver on port %d; id %d", port, me)
	select {}
}
