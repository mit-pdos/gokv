package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/mit-pdos/gokv/grove_ffi"
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

func StartServer(hostname grove_ffi.Address) {
	ln := grove_ffi.Listen(hostname)

	for {
		conn := grove_ffi.Accept(ln)

		go func() {
			for {
				r := grove_ffi.Receive(conn)
				if r.Err != false {
					return
					// panic("error while receiving")
				}

				go func() {
					grove_ffi.Send(conn, r.Data)
				}()
			}
		}()
	}
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	portnum := 8082
	go StartServer(grove_ffi.MakeAddress(fmt.Sprintf("127.0.0.1:%d", portnum)))
	log.Printf("Started echo grove_ffi server on port 8082")
	select {}
}
