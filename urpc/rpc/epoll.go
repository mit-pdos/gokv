package rpc

import (
	"net"
	"syscall"
	"reflect"
)

// not concurrency-safe; must only call Add(), Remove() with ownership discipline
type Epoller struct {
	fd int
	Conns map[int]net.Conn
}

func MakeEpoller() *Epoller {
	fd, err := syscall.EpollCreate1(0)
	// fmt.Println(fd)
	if err != nil {
		panic(err)
	}
	e := new(Epoller)
	e.fd = fd
	e.Conns = make(map[int]net.Conn)
	return e
}

func (e *Epoller) Wait() []syscall.EpollEvent {
	events := make([]syscall.EpollEvent, 50)
retry:
	n, err := syscall.EpollWait(e.fd, events, -1)
	if err == syscall.EINTR {
		// fmt.Println("EINTR")
		goto retry
	}
	if err != nil {
		panic(err)
	}
	return events[:n]
}

func (e *Epoller) Add(c net.Conn) {
	fd := socketFD(c)
	err := syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Events: syscall.EPOLLIN, Fd: int32(fd)})
	if err != nil {
		panic(err)
	}
	e.Conns[fd] = c
}

func (e *Epoller) Remove(c net.Conn) {
	fd := socketFD(c)
	err := syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		panic(err)
	}
	delete(e.Conns, fd)
}

// XXX: from https://github.com/smallnest/1m-go-tcp-server/blob/master/2_epoll_server/epoll_linux.go
func socketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}
