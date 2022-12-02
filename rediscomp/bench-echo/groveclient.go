package main

import "github.com/mit-pdos/gokv/grove_ffi"

func groveInitClient() func() {
	connRet := grove_ffi.Connect(grove_ffi.MakeAddress(serverAddress))
	if connRet.Err != false {
		panic("error while connecting")
	}
	conn := connRet.Connection

	msg := make([]byte, msgSize)

	return func() {
		err := grove_ffi.Send(conn, msg)
		if err != false {
			panic("error while sending")
		}
		r := grove_ffi.Receive(conn)
		if r.Err != false {
			panic("error while receiving")
		} else if len(r.Data) != msgSize {
			panic("did not receive full message back")
		}
	}
}
