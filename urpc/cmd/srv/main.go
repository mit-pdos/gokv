package main

import (
	"github.com/upamanyus/urpc/rpc"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(4)
	fmt.Println("Starting server on port 12345")
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[1] = func(args []byte, reply *[]byte) {
		*reply = []byte("This works!")
		return
	}
	handlers[2] = func(args []byte, reply *[]byte) {
		*reply = make([]byte, 16)
		return
	}
	s := rpc.MakeRPCServer(handlers, 2)
	s.Serve(":12345")
	select{}
}
