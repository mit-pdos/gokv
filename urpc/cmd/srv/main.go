package main

import (
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
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
	s := urpc.MakeServer(handlers)
	s.Serve(grove_ffi.MakeAddress("127.0.0.1:12345"))
	select {}
}
