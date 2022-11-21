package main

import (
	"fmt"
	"log"
	"net"
)

//
// Runs an "echo" server that just receives "messages" over TCP and returns
// whatever it got. A "message" is assumed to fit in one `conn.Read()`.
//

func panic_if_err(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("Error: %s\nMessage: %s", err.Error(), msg))
	}
}

func StartServer(portnum int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", portnum))
	panic_if_err(err, "Listen() failed:")

	for {
		conn, err := ln.Accept()
		panic_if_err(err, "Listen")

		go func() {
			for {
				// XXX: need to allocate buffer in the loop because the replying
				// goroutine uses it.
				buffer := make([]byte, 128)

				n, err := conn.Read(buffer)
				if err != nil {
					return
				}

				go func() {
					conn.Write(buffer[:n])
				}()
			}
		}()
	}
}

func main() {
	go StartServer(8080)
	log.Printf("Started echo server on port 8080")
	select {}
}
