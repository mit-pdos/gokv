package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/kvservice"
)

func main() {
	var confStr string
	flag.StringVar(&confStr, "host", "", "Address of kv server")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" put key value")
			fmt.Println(" cput key expValue value")
			fmt.Println(" get key")
			os.Exit(1)
		}
	}

	if len(confStr) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf := grove_ffi.MakeAddress(confStr)
	ck := kvservice.MakeClerk(conf)

	a := flag.Args()
	usage_assert(len(a) > 0)
	if a[0] == "put" {
		usage_assert(len(a) == 3)
		k := a[1]
		v := a[2]
		ck.Put(k, v)
		fmt.Printf("PUT %v ↦ %v\n", k, v)
	} else if a[0] == "cput" {
		usage_assert(len(a) == 4)
		k := a[1]
		ev := a[2]
		v := a[3]
		ok := ck.ConditionalPut(k, ev, v)
		if ok {
			fmt.Printf("CPUT %v ↦ %v ⤳ %v\n", k, ev, v)
		} else {
			fmt.Printf("CPUT failed %v ↦ %v\n", k, v)
		}
	} else if a[0] == "get" {
		usage_assert(len(a) == 2)
		k := a[1]
		v := ck.Get(k)
		fmt.Printf("GET %v ↦ %v\n", k, v)
	} else {
		usage_assert(false)
	}
}
