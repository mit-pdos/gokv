package rpc

import (
	"github.com/tchajed/marshal"
	"io"
	"net"
)

type MsgAndSender struct {
	m []byte
	s *Sender
}

/// Sender
type Sender struct {
    conn net.Conn
}


/// Receiver
type Receiver struct {
    c chan MsgAndSender
}

func MakeSender(host string) (*Sender, *Receiver) {
	// FIXME: cache "conn" in some global map to make connections live longer
	conn, err := net.Dial("tcp", host)
	// We ignore errors (all packets are just silently dropped)
	if err != nil { // keeping this around so it's easier to debug code
		panic(err)
	}
	c := make(chan MsgAndSender)
	go receiveOnSocket(conn, c)
	return &Sender { conn }, &Receiver { c }
}

func Send(send *Sender, data []byte) {
	e := marshal.NewEnc(8 + uint64(len(data)))
	e.PutInt(uint64(len(data)))
	e.PutBytes(data)
	reqData := e.Finish()
	send.conn.Write(reqData) // one atomic write for the entire thing!
}

func receiveOnSocket(conn net.Conn, c chan MsgAndSender) {
	for {
		// reply format: [dataLen] ++ data
		header := make([]byte, 8)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			panic(err)
		}
		d := marshal.NewDec(header)
		dataLen := d.GetInt()

		data := make([]byte, dataLen)
		_, err2 := io.ReadFull(conn, data)
		if err2 != nil {
			panic(err2)
		}
		c <- MsgAndSender{data, &Sender{conn}}
	}
}

func listenOnSocket(l net.Listener, c chan MsgAndSender) {
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err) // Here for easier debugging
		}
		// Spawn new thread receiving data on this connection
		go receiveOnSocket(conn, c)
	}
}

func MakeReceiver(host string) *Receiver {
	c := make(chan MsgAndSender)
	l, err := net.Listen("tcp", host)
	if err != nil {
		return &Receiver { c }
	}
	// Keep accepting new connections in background thread
	go listenOnSocket(l, c)
	return &Receiver { c }
}

// This will never actually return NULL, but as long as clients and proofs do not rely on this that is okay.
func Receive(recv *Receiver) ([]byte, *Sender) {
	a := <-recv.c
	return a.m, a.s
}
