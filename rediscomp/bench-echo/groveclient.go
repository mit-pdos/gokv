package main

import "github.com/mit-pdos/gokv/grove_ffi"

func groveInitClient() func() {
	err, conn := grove_ffi.Connect(grove_ffi.MakeAddress(serverAddress))
	if err != false {
		panic("error while connecting")
	}

	msg := make([]byte, msgSize)

	return func() {
		err := grove_ffi.Send(conn, msg)
		if err != false {
			panic("error while sending")
		}
		err, data := grove_ffi.Receive(conn)
		if err != false {
			panic("error while receiving")
		} else if len(data) != msgSize {
			panic("did not receive full message back")
		}
	}
}
