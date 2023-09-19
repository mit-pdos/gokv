package main

import (
	"flag"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/vrsm/paxos"
	"os"
)

func main() {
	var host string
	flag.StringVar(&host, "host", "", "address of paxos server (e.g. 10.0.0.1:12345)")
	flag.Parse()

	if host == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	ck := mpaxos.MakeSingleClerk(grove_ffi.MakeAddress(host))
	ck.TryBecomeLeader()
}
