package main

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/paxi/reconf"
	"log"
)

func main() {
	srvStrings := []string{
		"127.0.0.1:12200",
		"127.0.0.1:12201",
		"127.0.0.1:12202",
		"127.0.0.1:12203",
		"127.0.0.1:12204",
	}
	srvs := make([]grove_ffi.Address, len(srvStrings))

	for i := range srvs {
		srvs[i] = grove_ffi.MakeAddress(srvStrings[i])
		log.Println(srvs[i])
	}

	initConfig := &reconf.Config{Members: srvs[:3]}

	for _, addr := range srvs {
		reconf.StartReplicaServer(addr, initConfig)
	}

	ck := reconf.MakeClerkPool()
	if ck.TryCommitVal(srvs[0], []byte("Hello!")) == false {
		log.Fatal("Unable to commit")
	}
}
