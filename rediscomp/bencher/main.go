package main

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/rediscomp/benchclosed"
	"github.com/mit-pdos/gokv/urpc"
)

var msgSize int
var serverAddress string

func startClientConnection(done *uint64, pauseDuration time.Duration) {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		panic(err)
	}
	go generateLoad(conn, pauseDuration)
	go receiveEchos(done, conn)
}

func generateLoad(conn net.Conn, pauseDuration time.Duration) {
	msg := make([]byte, msgSize)
	for {
		n, err := conn.Write(msg)
		time.Sleep(pauseDuration)

		if err != nil {
			panic(err)
		} else if n != msgSize {
			panic("Write didn't write the whole message")
		}
	}
}

func receiveEchos(donePtr *uint64, conn net.Conn) int {
	numMessagesReceived := 0

	msg := make([]byte, msgSize)

	for atomic.LoadUint64(donePtr) == 0 {
		n, err := conn.Read(msg)
		if err != nil {
			panic(err)
		} else if n != msgSize {
			panic("Read didn't return the whole message")
		}

		numMessagesReceived += 1
	}

	fmt.Printf("Received: %d\n", numMessagesReceived)
	return numMessagesReceived
}

func echoInitClient() func() {
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

func main() {
	msgSize = 128
	serverAddress = "127.0.0.1:8080"
	numClients := 50
	warmup := 1 * time.Second
	runtime := 10 * time.Second

	benchclosed.RunBench(urpcInitClient,
		numClients,
		runtime,
		warmup,
	)
}
