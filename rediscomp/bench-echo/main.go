package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/rediscomp/benchclosed"
	"github.com/mit-pdos/gokv/urpc"
)

var msgSize int
var serverAddress string

func netInitClient() func() {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		panic(err)
	}
	msg := make([]byte, msgSize)

	return func() {
		n, err := conn.Write(msg)
		if err != nil {
			panic(err)
		} else if n != msgSize {
			panic("Write didn't write the whole message")
		}

		n, err = conn.Read(msg)
		if err != nil {
			panic(err)
		} else if n != msgSize {
			panic("Read didn't return the whole message")
		}
	}
}

func urpcInitClient() func() {
	cl := urpc.MakeClient(grove_ffi.MakeAddress(serverAddress))
	args := make([]byte, msgSize)
	reply := new([]byte)

	return func() {
		cl.Call(0, args, reply, 100 /* ms */)
	}
}

func usage() {
	fmt.Println("Must provide indicate which server to run: urpc | net")
}

func main() {
	msgSize = 128
	numClients := 50
	var runtimeSeconds = flag.Int("runtime", 10, "number of seconds to run benchmark for")

	var warmupSeconds = flag.Int("warmup", 1, "number of seconds to warm up before measuring")
	flag.Parse()

	runtime := time.Duration(*runtimeSeconds) * time.Second
	warmup := time.Duration(*warmupSeconds) * time.Second

	args := flag.Args()
	if len(args) != 1 {
		usage()
		return
	}

	initClient := urpcInitClient

	// use different addresses for the two servers to avoid accidentally
	// connecting to the wrong server
	if args[0] == "net" {
		initClient = netInitClient
		serverAddress = "127.0.0.1:8080"
	} else if args[0] == "urpc" {
		initClient = urpcInitClient
		serverAddress = "127.0.0.1:8081"
	} else if args[0] == "grove" {
		initClient = groveInitClient
		serverAddress = "127.0.0.1:8082"
	} else {
		usage()
		panic("invalid command provided")
	}

	benchclosed.RunBench(initClient,
		numClients,
		warmup,
		runtime,
	)
}
