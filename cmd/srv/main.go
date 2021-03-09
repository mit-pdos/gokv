package main

import (
	"fmt"
	"github.com/upamanyus/gokv"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
	"os"
	"runtime/pprof"
	"flag"
)
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	srv := gokv.MakeGoKVServer()

	flag.Parse()
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

	rpc.RegisterName("KV", srv)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":12345")
	if e != nil {
		panic(e)
	}

	fmt.Println("Starting server")
	// go http.Serve(l, nil)
	func() {
		log.Fatal(http.Serve(l, nil))
	}()
	time.Sleep(10 * time.Second)
}
