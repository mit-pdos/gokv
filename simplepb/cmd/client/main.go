package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/client"
)

func main() {
	var confStr string
	flag.StringVar(&confStr, "conf", "", "Address of configuration server")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" faa key value")
			fmt.Println(" get key")
			os.Exit(1)
		}
	}

	if len(confStr) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf := grove_ffi.MakeAddress(confStr)
	ck := client.Make(conf)

	a := flag.Args()
	usage_assert(len(a) > 0)
	if a[0] == "faa" {
		usage_assert(len(a) == 3)
		k, err := strconv.ParseUint(a[1], 10, 64)
		usage_assert(err == nil)

		v := []byte(a[2])
		oldv := ck.FetchAndAppend(k, v)
		fmt.Printf("FAA %d %v, old was %v\n", k, v, oldv)
	} else if a[0] == "get" {
		usage_assert(len(a) == 2)
		k, err := strconv.ParseUint(a[1], 10, 64)
		usage_assert(err == nil)

		v := ck.FetchAndAppend(k, make([]byte, 0))
		fmt.Printf("GET %d ↦ %v\n", k, v)
	}
}
