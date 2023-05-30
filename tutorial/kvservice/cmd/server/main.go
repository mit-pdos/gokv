package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/kvservice"
)

func main() {
	// for performance profiling
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.Parse()
	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	kvservice.MakeServer().Start(me)
	log.Printf("Started kv server on port %d; id %d", port, me)
	select {}
}
