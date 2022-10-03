package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/kv"
	"log"
	"os"
)

func main() {
	var fname string
	var port uint64
	flag.StringVar(&fname, "filename", "", "name of file that holds durable state for this server")
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.Parse()

	if fname == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	kv.Start(fname, me)
	log.Printf("Started kv server on port %d; id %d", port, me)
	select {}
}
