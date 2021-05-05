package dist_ffi

import (
	"fmt"
	"github.com/tchajed/marshal"
	"io"
	"net"
	"strconv"
	"strings"
)

type Address uint64

func (a Address) String() string {
	return AddressToStr(a)
}

func MakeAddress(ipStr string) uint64 {
	// XXX: manually parsing is pretty silly; couldn't figure out how to make
	// this work cleanly net.IP
	ipPort := strings.Split(ipStr, ":")
	if len(ipPort) != 2 {
		panic(fmt.Sprintf("Not ipv4:port %s", ipStr))
	}
	port, err := strconv.ParseUint(ipPort[1], 10, 16)
	if err != nil {
		panic(err)
	}

	ss := strings.Split(ipPort[0], ".")
	if len(ss) != 4 {
		panic(fmt.Sprintf("Not ipv4:port %s", ipStr))
	}
	ip := make([]byte, 4)
	for i, s := range ss {
		a, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			panic(err)
		}
		ip[i] = byte(a)
	}
	return (uint64(ip[0]) | uint64(ip[1])<<8 | uint64(ip[2])<<16 | uint64(ip[3])<<24 | uint64(port)<<32)
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
	return fmt.Sprintf("%s:%d", net.IPv4(a0, a1, a2, a3).String(), port)
}

type MsgAndSender struct {
	m []byte
	s Sender
}

/// Sender
type sender struct {
	conn net.Conn
	host Address // to reconnect to server when it fails (can be 0 to indicate this is not possible)
	recv Receiver // the reply receiver for this channel
}

type Sender *sender

type ConnectRet struct {
	Err      bool
	Sender   Sender
	Receiver Receiver
}

func Connect(host Address) ConnectRet {
	conn, err := net.Dial("tcp", AddressToStr(host))
	c := make(chan MsgAndSender)
	recv := &receiver{c}
	send := &sender { conn, host, recv }

	if err == nil {
		go receiveOnSocket(conn, c)
	}
	return ConnectRet{Err: err != nil, Sender: send, Receiver: recv}
}

func Send(send Sender, data []byte) bool {
	// message format: [dataLen] ++ data
	e := marshal.NewEnc(8 + uint64(len(data)))
	e.PutInt(uint64(len(data)))
	e.PutBytes(data) // FIXME: copying all the data...
	reqData := e.Finish()
	_, err := send.conn.Write(reqData) // one atomic write for the entire thing!
	if err != nil && send.host != 0 {
		// This did not work out. In an attempt to make this API as reliable as possible,
		// let us try to reconnect so if the client tries again, maybe it works.
		conn, err := net.Dial("tcp", AddressToStr(send.host))
		if err == nil {
			// Looking good, we got a new connection. Let's use this henceforth and
			// wire it up to the existing receiver's channel.
			send.conn = conn
			go receiveOnSocket(conn, send.recv.c)
		}
	}
	return err != nil
}

// conn will also be used as "reply socket" for all messages that arrive here.
func receiveOnSocket(conn net.Conn, c chan MsgAndSender) {
	// Messages received here will have their replies sent via this sender.
	// "host" and "recv" remain zero; this sender does not support re-connecting.
	var send sender
	send.conn = conn

	for {
		// message format: [dataLen] ++ data
		header := make([]byte, 8)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			// TODO: if this is a `Receiver`, propagate socket failures to `Receive` calls
			// (Hiding errors is okay per our spec, but not great.)
			// This can legitimately happen when the other side "hung up", so do not panic.
			return
		}
		d := marshal.NewDec(header)
		dataLen := d.GetInt()

		data := make([]byte, dataLen)
		_, err2 := io.ReadFull(conn, data)
		if err2 != nil {
			// see comment above
			return
		}
		c <- MsgAndSender{data, &send}
	}
}

/// Receiver
type receiver struct {
	c chan MsgAndSender
}

type Receiver *receiver

func listenOnSocket(l net.Listener, c chan MsgAndSender) {
	for {
		conn, err := l.Accept()
		if err != nil {
			// Assume() no error on Listen
			panic(err)
		}
		// Spawn new thread receiving data on this connection
		go receiveOnSocket(conn, c)
	}
}

func Listen(host Address) Receiver {
	c := make(chan MsgAndSender)
	l, err := net.Listen("tcp", AddressToStr(host))
	if err != nil {
		// Assume() no error on Listen
		panic(err)
	}
	// Keep accepting new connections in background thread
	go listenOnSocket(l, c)
	return &receiver{c}
}

type ReceiveRet struct {
	Err    bool
	Sender Sender
	Data   []byte
}

// This will never actually return Err==true, but as long as clients and proofs do not rely on this that is okay.
// TODO: Distinguish "timeout (no message found)" from "error"
func Receive(recv Receiver) ReceiveRet {
	a := <-recv.c
	return ReceiveRet{Err: false, Sender: a.s, Data: a.m}
}
