package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/fastkv"
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

	gkv := fastkv.MakeGoKVShardServer()
	gkv.Start()

	fmt.Println("Started FastKV server")
	select {} // sleep forever
}
