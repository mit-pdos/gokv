package main

import (
	"github.com/mit-pdos/gokv/memkv"
	"fmt"
	"flag"
	// "log"
	// "os"
)

func main() {
	var coord string
	flag.StringVar(&coord, "coord", "", "address of coordinator")
	flag.Parse()

	if coord == "" {
		// flag.PrintDefaults()
		// os.Exit(1)
	}

	coord = "localhost:37000"
	ck := memkv.MakeMemKVClerk(coord)
	ck.Put(15, []byte("This is a test"))
	fmt.Printf("Got: %s", string(ck.Get(15)))
}
