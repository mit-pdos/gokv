package rpc

import (
	"github.com/tchajed/marshal"
	"io"
	"net"
)

/// Sender
type Sender struct {
    conn net.Conn
}

func MakeSender(host string) *Sender {
	// FIXME: cache "conn" in some global map to make connections live longer
	conn, _ := net.Dial("tcp", host)
	// We ignore errors (all packets are just silently dropped)
	return &Sender { conn }
}

func Send(send *Sender, data []byte) {
	e := marshal.NewEnc(8 + uint64(len(data)))
	e.PutInt(uint64(len(data)))
	e.PutBytes(data)
	reqData := e.Finish()
	send.conn.Write(reqData) // one atomic write for the entire thing!
}

/// Receiver
type Receiver struct {
    c chan []byte
}

func receiveOnSocket(conn net.Conn, c chan []byte) {
	for {
		// reply format: [dataLen] ++ data
		header := make([]byte, 8)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			return
		}
		d := marshal.NewDec(header)
		dataLen := d.GetInt()

		data := make([]byte, dataLen)
		_, err2 := io.ReadFull(conn, data)
		if err2 != nil {
			return
		}
		c <- data
	}
}

func listenOnSocket(l net.Listener, c chan []byte) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		// Spawn new thread receiving data on this connection
		go receiveOnSocket(conn, c)
	}
}

func MakeReceiver(host string) *Receiver {
	c := make(chan []byte)
	l, err := net.Listen("tcp", host)
	if err != nil {
		return &Receiver { c }
	}
	// Keep accepting new connections in background thread
	go listenOnSocket(l, c)
	return &Receiver { c }
}

// This will never actually return NULL, but as long as clients and proofs do not rely on this that is okay.
func Receive(recv *Receiver) []byte {
	return <-recv.c
}
