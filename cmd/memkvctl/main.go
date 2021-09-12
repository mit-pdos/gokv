package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/memkv"
	"os"
	"strconv"
)

func main() {
	var coordStr string
	flag.StringVar(&coordStr, "coord", "", "address of coordinator")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" get KEY")
			fmt.Println(" put KEY VALUE")
			fmt.Println(" add HOST")
			os.Exit(1)
		}
	}

	usage_assert(coordStr != "")

	coord := grove_ffi.MakeAddress(coordStr)
	ck := memkv.MakeMemKVClerk(coord, connman.MakeConnMan())

	a := flag.Args()
	usage_assert(len(a) > 0)
	if a[0] == "get" {
		usage_assert(len(a) == 2)
		k, err := strconv.ParseUint(a[1], 10, 64)
		usage_assert(err == nil)
		v := ck.Get(k)
		fmt.Printf("GET %d ↦ %v\n", k, v)
	} else if a[0] == "put" {
		usage_assert(len(a) == 3)
		k, err := strconv.ParseUint(a[1], 10, 64)
		usage_assert(err == nil)
		v := []byte(a[2])
		ck.Put(k, v)
		fmt.Printf("PUT %d ↦ %v\n", k, v)
	} else if a[0] == "add" {
		usage_assert(len(a) == 2)
		h := grove_ffi.MakeAddress(a[1])
		ck.Add(h)
		fmt.Printf("Added %s\n", a[1])
	}
}
