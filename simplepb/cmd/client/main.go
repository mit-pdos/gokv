package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/kv"
)

func main() {
	var confStr string
	flag.StringVar(&confStr, "conf", "", "Address of configuration server")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" put key value")
			fmt.Println(" get key")
			os.Exit(1)
		}
	}

	if len(confStr) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf := grove_ffi.MakeAddress(confStr)
	ck := kv.MakeClerk(conf)

	a := flag.Args()
	usage_assert(len(a) > 0)
	if a[0] == "put" {
		usage_assert(len(a) == 3)
		ck.Put(a[1], a[2])
		fmt.Printf("PUT %s ↦ %s\n", a[1], a[2])
	} else if a[0] == "get" {
		usage_assert(len(a) == 2)
		v := ck.Get(a[1])
		fmt.Printf("GET %s ↦ %s\n", a[1], v)
	} else {
		usage_assert(false)
	}
}
