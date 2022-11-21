package main

import (
	"log"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

func main() {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[0] = func(args []byte, reply *[]byte) {
		*reply = args
	}
	s := urpc.MakeServer(handlers)
	s.Serve(grove_ffi.MakeAddress("127.0.0.1:8080"))
	log.Println("Started echo-urpc server on port 8080")
	select {}
}
