package main

import (
	"fmt"
	"github.com/felixge/fgprof"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"log"
	"net/http"
	_ "net/http/pprof"
)

const (
	RPC_NULL = uint64(0) // equal to KV_FRESHCID
	KV_GET   = uint64(2)
)

func main() {
	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_NULL] = func(args []byte, reply *[]byte) {
		e := marshal.NewEnc(8)
		e.PutInt(0)
		*reply = e.Finish()
	}
	handlers[KV_GET] = func(args []byte, reply *[]byte) {
		e := marshal.NewEnc(8 + 8)
		e.PutInt(0)
		e.PutInt(0)
		*reply = e.Finish()
	}

	s := rpc.MakeRPCServer(handlers)
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", 12345))
	log.Printf("Started null RPC server on port %d; id %d", 12345, me)
	s.Serve(me, 1)
	select {}
}
