package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/memkv"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// var coord string
	var is_init bool
	var port uint64
	flag.BoolVar(&is_init, "init", false, "true iff this server owns all shard at initialization; default is false")
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	// flag.StringVar(&coord, "coord", "", "address of coordinator")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := memkv.MakeKVShardServer(is_init)
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	log.Printf("Started shard server on port %d; id %d", port, me)
	s.Start(me)
	select {}
}
