package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/pb/controller"
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

	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	primary := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:12202"))
	replicas := make([]uint64, 0)
	replicas = append(replicas, primary)
	replicas = append(replicas, grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:12203")))
	replicas = append(replicas, grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:12204")))
	replicas = append(replicas, grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:12205")))
	replicas = append(replicas, grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:12206")))
	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	log.Printf("Connecting to id %d", primary)
	controller.StartControllerServer(me, replicas)
	log.Printf("Started controller server on port %d; id %d", port, me)

	select {}
}
