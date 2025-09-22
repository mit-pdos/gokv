package lockservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/lockservice/lockrequest_gk"
	"github.com/mit-pdos/gokv/urpc"
)

const (
	RPC_GET_FRESH_NUM = uint64(0)
	RPC_TRY_ACQUIRE   = uint64(1)
	RPC_RELEASE       = uint64(2)
)

type Error = uint64

type Client struct {
	cl *urpc.Client
}

func (cl *Client) getFreshNum() (uint64, Error) {
	var reply []byte
	args := make([]byte, 0)
	err := cl.cl.Call(RPC_GET_FRESH_NUM, args, &reply, 100)
	if err == urpc.ErrNone {
		return DecodeUint64(reply), err
	}
	return 0, err
}

func (cl *Client) tryAcquire(id uint64) (uint64, Error) {
	var reply []byte
	args := lockrequest_gk.Marshal([]byte{}, lockrequest_gk.S{Id: id})
	err := cl.cl.Call(RPC_TRY_ACQUIRE, args, &reply, 100)
	if err == urpc.ErrNone {
		return DecodeUint64(reply), err
	}
	return 0, err
}

func (cl *Client) release(id uint64) Error {
	var reply []byte
	args := lockrequest_gk.Marshal([]byte{}, lockrequest_gk.S{Id: id})
	return cl.cl.Call(RPC_RELEASE, args, &reply, 100)
}

func makeClient(hostname grove_ffi.Address) *Client {
	return &Client{cl: urpc.MakeClient(hostname)}
}
