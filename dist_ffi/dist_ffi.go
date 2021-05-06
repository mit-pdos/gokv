package dist_ffi

import (
	"fmt"
	"github.com/tchajed/marshal"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
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
	mu *sync.Mutex
	host Address //[read-only] to reconnect to server when it fails (can be 0 to indicate this is not possible)
	replies chan MsgAndSender //[read-only] the channel on which replies arrive
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
	send := &sender { conn: conn, host: host, replies: c, mu: new(sync.Mutex) }

	// On an error we lost the connection -- so close the channel to tell the Receiver.
	// FIXME: this is wrong and can lead to panics, we reconnect and there might be other `receiveOnSocket`!
	go receiveOnSocket(send, /*close_on_err*/true)
	return ConnectRet{Err: err != nil, Sender: send, Receiver: &receiver{c}}
}

func Send(send Sender, data []byte) bool {
	// message format: [dataLen] ++ data
	e := marshal.NewEnc(8)
	e.PutInt(uint64(len(data)))
	reqLen := e.Finish()

	send.mu.Lock() // ensure the two writes and the potential reconnect are "atomic"
	_, err := send.conn.Write(reqLen)
	if err == nil {
		_, err = send.conn.Write(data)
	}
	if err != nil {
		// This socket is broken, make sure we do not send anything on it ever again.
		send.conn.Close()
		if send.host != 0 {
			// In an attempt to make this API as reliable as possible,
			// let us try to reconnect so when the client tries again, maybe it works.
			conn, err := net.Dial("tcp", AddressToStr(send.host))
			if err == nil {
				// Looking good, we got a new connection. Let's use this henceforth.
				send.conn = conn
				// On an error we lost the connection -- so close the channel to tell the Receiver.
				// FIXME: this is wrong and can lead to panics, we reconnect and there might be other `receiveOnSocket`!
				go receiveOnSocket(send, /*close_on_err*/true)
			}
		}
	}
	send.mu.Unlock()
	return err != nil
}

// Handle the receive direction of the given sender.
// `close_on_err` indicates if on an error, the channel should be closed.
func receiveOnSocket(send Sender, close_on_err bool) {
	for {
		// message format: [dataLen] ++ data
		header := make([]byte, 8)
		_, err := io.ReadFull(send.conn, header)
		if err != nil {
			// Looks like this connection is dead.
			// This can legitimately happen when the other side "hung up", so do not panic.
			if close_on_err {
				close(send.replies)
			}
			return
		}
		d := marshal.NewDec(header)
		dataLen := d.GetInt()

		data := make([]byte, dataLen)
		_, err = io.ReadFull(send.conn, data)
		if err != nil {
			// Connection interrupted after some of the data was sent, or so...
			if close_on_err {
				close(send.replies)
			}
			return
		}
		send.replies <- MsgAndSender{data, send}
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
			// This should not usually happen... something seems wrong.
			panic(err)
		}
		// Spawn new thread receiving data on this connection.
		send := &sender {
			conn: conn,
			mu: new(sync.Mutex),
			replies: c,
			host: 0, // do support for reconnecting
		}
		// Errors just mean one client disappeared, do not close the channel.
		go receiveOnSocket(send, /*close_on_err*/false)
	}
}

func Listen(host Address) Receiver {
	c := make(chan MsgAndSender)
	l, err := net.Listen("tcp", AddressToStr(host))
	if err != nil {
		// Assume() no error on Listen. This should fail loud and early, retrying makes little sense (likely the port is already used).
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

func Receive(recv Receiver) ReceiveRet {
	a, more := <-recv.c
	return ReceiveRet{Err: !more, Sender: a.s, Data: a.m}
}
