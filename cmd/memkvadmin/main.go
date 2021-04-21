package main

import (
	"github.com/mit-pdos/gokv/memkv"
	"github.com/mit-pdos/gokv/dist_ffi"
	"flag"
)

func main() {
	var coordStr string
	flag.StringVar(&coordStr, "coord", "", "address of coordinator")
	flag.Parse()

	if coordStr == "" {
		// flag.PrintDefaults()
		// os.Exit(1)
	}

	coordStr = "127.0.0.1:37000"
	coord := dist_ffi.MakeAddress(coordStr)
	h := dist_ffi.MakeAddress("127.0.0.1:37001")
	ck := memkv.MakeMemKVClerk(coord)
	ck.Add(h)
	// ck.Put(15, []byte("This is a test"))
	// fmt.Printf("Got: %s", string(ck.Get(15)))
}
