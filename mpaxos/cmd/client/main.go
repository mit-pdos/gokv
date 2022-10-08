package main

import (
	"flag"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/mpaxos"
)

func main() {
	flag.Parse()
	a := flag.Args()

	config := make([]grove_ffi.Address, len(a))
	for i := range config {
		config[i] = grove_ffi.MakeAddress(a[i])
	}

	ck := mpaxos.MakeClerk(config)
	ck.Apply(make([]byte, 0))
}
