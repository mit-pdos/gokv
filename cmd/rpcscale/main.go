package main

import (
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"log"
	"net/http"
	_ "net/http/pprof"
)

const (
	RPC_NULL = iota
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_NULL] = func([]byte, *[]byte) {
	}

	s := rpc.MakeRPCServer(handlers)
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", 12345))
	log.Printf("Started null RPC server on port %d; id %d", 12345, me)
	s.Serve(me, 1)
	select {}
}
