package main

import (
	"fmt"
	"flag"
	"github.com/mit-pdos/gokv/goosekv"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	gkv := goosekv.MakeGoKVShardServer()
	grove_ffi.SetPort(12345)
	gkv.Start()

	fmt.Println("Started GooseKV server")
	select{} // sleep forever
}
