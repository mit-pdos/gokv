package dist_ffi

import (
	"github.com/tchajed/marshal"
	"io"
	"net"
	"fmt"
	"strings"
	"strconv"
)

type Address uint64

func MakeAddress(ipStr string, port uint16) uint64 {
	// XXX: manually parsing is pretty silly; couldn't figure out how to make
	// this work cleanly net.IP
	ss := strings.Split(ipStr, ".")
	if len(ss) != 4 {
		panic(fmt.Sprintf("Not ipv4 %s", ipStr))
	}
	ip := make([]byte, 4)
	for i, s := range ss {
		a, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			panic(err)
		}
		ip[i] = byte(a)
	}
	return (uint64(ip[0]) | uint64(ip[1]) << 8 | uint64(ip[2]) << 16 | uint64(ip[3]) << 24 | uint64(port) << 32)
}

func AddressToStr(e Address) string {
	a0 := byte(e & 0xff)
	e = e >> 8
	a1 := byte(e & 0xff)
	e = e >> 8
	a2 := byte(e & 0xff)
	e = e >> 8
	a3 := byte(e & 0xff)
	e = e >> 8
	port := e & 0xffff
	return fmt.Sprintf("%s:%d", net.IPv4(a0,a1,a2,a3).String(), port)
}

type MsgAndSender struct {
	m []byte
	s *Sender
}

/// Sender
type Sender struct {
    conn net.Conn
}

type SenderReceiver struct {
	S *Sender
	R *Receiver
}

func Connect(host Address) SenderReceiver {
	conn, err := net.Dial("tcp", AddressToStr(host))
	// We ignore errors (all packets are just silently dropped)
	if err != nil { // keeping this around so it's easier to debug code
		panic(err)
	}
	c := make(chan MsgAndSender)
	go receiveOnSocket(conn, c)
	return SenderReceiver { S:&Sender { conn }, R:&Receiver { c } }
}

func Send(send *Sender, data []byte) {
	// message format: [dataLen] ++ data
	e := marshal.NewEnc(8 + uint64(len(data)))
	e.PutInt(uint64(len(data)))
	e.PutBytes(data) // FIXME: copying all the data...
	reqData := e.Finish()
	send.conn.Write(reqData) // one atomic write for the entire thing!
}

func receiveOnSocket(conn net.Conn, c chan MsgAndSender) {
	for {
		// message format: [dataLen] ++ data
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
			panic(err2)
		}
		c <- MsgAndSender{data, &Sender{conn}}
	}
}

/// Receiver
type Receiver struct {
    c chan MsgAndSender
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

func Listen(host Address) *Receiver {
	c := make(chan MsgAndSender)
	l, err := net.Listen("tcp", AddressToStr(host))
	if err != nil {
		return &Receiver { c }
	}
	// Keep accepting new connections in background thread
	go listenOnSocket(l, c)
	return &Receiver { c }
}

type ErrMsgSender struct {
	E bool
	M []byte
	S *Sender
}

// This will never actually return NULL, but as long as clients and proofs do not rely on this that is okay.
func Receive(recv *Receiver) ErrMsgSender {
	a := <-recv.c
	return ErrMsgSender{E:false, M:a.m, S:a.s}
}
