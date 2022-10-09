package main

import (
	"flag"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/mpaxos/example"
)

func main() {
	flag.Parse()
	a := flag.Args()

	config := make([]grove_ffi.Address, 0)
	for i := range config {
		config[i] = grove_ffi.MakeAddress(a[i])
	}

	/*
	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" put value")
			fmt.Println(" get")
			os.Exit(1)
		}
	}
	*/

	ck := example.MakeClerk(config)
	ck.Get()
}
