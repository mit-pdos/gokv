package dist_ffi

import (
	"time"
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
	mu *sync.Mutex // to serialize writing to `conn`
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
	send := &sender { conn: conn, mu: new(sync.Mutex) }

	if err == nil {
		go receiveOnSocket(send, c)
	}
	return ConnectRet{Err: err != nil, Sender: send, Receiver: recv}
}

func Send(send Sender, data []byte) bool {
	// Encode length
	e := marshal.NewEnc(8)
	e.PutInt(uint64(len(data)))
	reqLen := e.Finish()
	// message format: [dataLen] ++ data
	send.mu.Lock()
	_, err := send.conn.Write(reqLen)
	if err == nil {
		_, err = send.conn.Write(data)
	}
	// If there was an error, make sure we never send anything on this channel again...
	// there might have been a partial write!
	if err != nil {
		send.conn.Close()
	}
	send.mu.Unlock()
	return err != nil
}

// Handle the receiving side of the given sender, and put the messages onto channel `c`.
func receiveOnSocket(send Sender, c chan MsgAndSender) {
	for {
		// message format: [dataLen] ++ data
		header := make([]byte, 8)
		_, err := io.ReadFull(send.conn, header)
		if err != nil {
			// TODO: if this is a `Receiver`, propagate socket failures to `Receive` calls
			// (Hiding errors is okay per our spec, but not great.)
			// This can legitimately happen when the other side "hung up", so do not panic.
			return
		}
		d := marshal.NewDec(header)
		dataLen := d.GetInt()

		data := make([]byte, dataLen)
		_, err2 := io.ReadFull(send.conn, data)
		if err2 != nil {
			// see comment above
			return
		}
		c <- MsgAndSender{data, send}
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
		// Spawn new thread receiving data on this connection
		send := &sender { conn: conn, mu: new(sync.Mutex) }
		go receiveOnSocket(send, c)
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
	Err    uint64 // 0 = success, 1 = timeout, 2 = other error
	Sender Sender
	Data   []byte
}

func Receive(recv Receiver, timeout_ms uint64) ReceiveRet {
	select {
	case <-time.After(time.Duration(timeout_ms * 1000 * 1000)): // convert to nanoseconds
		return ReceiveRet{Err: 1, Sender: nil, Data: nil}
	case a := <-recv.c:
		return ReceiveRet{Err: 0, Sender: a.s, Data: a.m}
	}
}
