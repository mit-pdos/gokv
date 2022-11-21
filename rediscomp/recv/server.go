package main

import (
	"fmt"
	"net"
	"time"
)

//
// Runs a "receive" server that just receives "messages" over TCP and tracks the
// number of messages received.
//

func panic_if_err(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("Error: %s\nMessage: %s", err.Error(), msg))
	}
}

func StartServer(portnum int) {
	ln, err := net.Listen("tcp", ":8080")
	panic_if_err(err, "Listen() failed:")

	for {
		conn, err := ln.Accept()
		panic_if_err(err, "Listen")

		go func() {
			var start time.Time
			var end time.Time

			warmupBytes := 1024
			totalBytes := warmupBytes + 1e6
			receivedBytes := 0

			buffer := make([]byte, 1024)

			// receive messages for a while
			for {
				n, err := conn.Read(buffer)
				panic_if_err(err, "")

				if receivedBytes < warmupBytes &&
					receivedBytes+n >= warmupBytes {
					start = time.Now()
				}
				receivedBytes += n

				if receivedBytes > totalBytes {
					end = time.Now()
					break
				}

				if n != 128 {
					panic("Expected 128 bytes")
				}
			}

			// XXX: this is approximate since when we set start = time.Now(), we
			// might have received more than warmupBytes bytes
			fmt.Printf("Took %v to receive ~%d bytes", end.Sub(start),
				receivedBytes-warmupBytes)
		}()
	}
}

func main() {
	panic("impl")
}
