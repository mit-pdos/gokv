package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/mit-pdos/gokv/rediscomp/redis"
)

const PROTO_IOBUF_LEN = 16 * 1024 // from server.h

type Server struct {
	mu  *sync.Mutex
	kvs map[string][]byte
}

func (s *Server) handleConnection(conn net.Conn) {
	buffer := make([]byte, PROTO_IOBUF_LEN)
	okReply := []byte("+OK\r\n")

	for {
		// https://redis.io/docs/reference/protocol-spec/
		// expecting "$<integer>\r\n"
		// Expecting to read full request.
		// This doesn't support pipelining or requests bigger than 16KB.

		n, err := conn.Read(buffer)
		if err != nil {
			return // done with conn
		}

		// parse command
		key, val := redis.ParseSetCommand(buffer[:n])
		// fmt.Printf("Setting %s -> %s\n", string(key), string(val))
		s.mu.Lock()
		s.kvs[string(key)] = val
		s.mu.Unlock()
		conn.Write(okReply)
	}
}

// returns key, value
func StartServer(portnum int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", portnum))
	if err != nil {
		panic(err)
	}

	s := &Server{
		mu:  new(sync.Mutex),
		kvs: make(map[string][]byte),
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go s.handleConnection(conn)
	}
}

func main() {
	go StartServer(6380)
	log.Println("Started miniredis net server")
	select {}
}
