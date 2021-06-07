package main

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/paxi/comulti"
	"log"
	"flag"
	// "time"
)

func main() {
	l := make([]comulti.Entry, 0)
	commitf := func(e comulti.Entry) {
		l = append(l, e)
		log.Printf("Log is %+v\n", l)
		if (len(l) % 100 == 0) {
			log.Println("Another 100")
		}
	}

	var i uint64
	flag.Uint64Var(&i, "index", 0, "the index of the server to start")

	var isLeader bool
	flag.BoolVar(&isLeader, "leader", false, "whether or now this node is initially leader")
	flag.Parse()

	peerHosts := []uint64{
		grove_ffi.MakeAddress("127.0.0.1:37001"),
		grove_ffi.MakeAddress("127.0.0.1:37002"),
		grove_ffi.MakeAddress("127.0.0.1:37003"),
	}

	r := comulti.MakeReplica(peerHosts[i], commitf, peerHosts, isLeader)
	log.Printf("Started replica")
	log.Printf("Started to try appending commands")
	for i := uint64(13); i < 1000000; i++ {
		go r.TryAppendRPC(i)
		go r.TryAppendRPC(i * 10)
		// time.Sleep(time.Millisecond * 500)
	}
	select {}
}
