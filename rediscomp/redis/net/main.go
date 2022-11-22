package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
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
		key, val := parseSetCommand(buffer[:n])
		// fmt.Printf("Setting %s -> %s\n", string(key), string(val))
		s.mu.Lock()
		s.kvs[string(key)] = val
		s.mu.Unlock()
		conn.Write(okReply)
	}
}

// returns key, value
func parseSetCommand(data []byte) ([]byte, []byte) {
	expectedPrefix := []byte("*3\r\n$3\r\nSET\r\n$")
	if !bytes.HasPrefix(data, expectedPrefix) {
		log.Fatalf("unexpected command; got %s", string(data))
	}
	data = data[len(expectedPrefix):]

	keyLen := 0

	// get size of key
	i := 0
	for data[i] != '\r' {
		keyLen = (keyLen * 10) + int(data[i]-'0')
		i++
	}

	if !(data[i+1] == '\n') {
		panic("expected LF")
	}
	data = data[i+2:]

	// now get the key
	if len(data) < keyLen+3 { // + 2 for \r\n$
		panic("incomplete SET command")
	}

	key := data[:keyLen]
	data = data[keyLen+3:]

	// get size of value
	i = 0
	valLen := 0
	for data[i] != '\r' {
		valLen = (valLen * 10) + int(data[i]-'0')
		i++
	}

	if !(data[i+1] == '\n') {
		panic("expected LF")
	}
	data = data[i+2:]

	val := data[:valLen]

	if !bytes.Equal(data[valLen:], []byte("\r\n")) {
		log.Fatalf("unexpected data after end of SET command; have %s", data[valLen:])
	}

	return key, val
}

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
