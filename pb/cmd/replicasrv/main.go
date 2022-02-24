package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/pb"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
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

	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	s := pb.StartReplicaServer(me)
	log.Printf("Started replica server on port %d; id %d", port, me)

	time.Sleep(2000 * time.Millisecond)
	for {
		s.StartAppend(byte(rand.Uint64() % 256))
		time.Sleep(500 * time.Millisecond)
		log.Printf("CommitLog[%d] :%+v\n", port, s.GetCommittedLog())
	}
	// select {}
}
