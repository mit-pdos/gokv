package kvservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/kvservice/conditionalput_gk"
	"github.com/mit-pdos/gokv/tutorial/kvservice/get_gk"
	"github.com/mit-pdos/gokv/tutorial/kvservice/put_gk"
	"github.com/mit-pdos/gokv/urpc"
)

const (
	rpcIdGetFreshNum    = uint64(0)
	rpcIdPut            = uint64(1)
	rpcIdConditionalPut = uint64(2)
	rpcIdGet            = uint64(3)
)

type Error = uint64

type Client struct {
	cl *urpc.Client
}

func (cl *Client) getFreshNumRpc() (uint64, Error) {
	var reply []byte
	err := cl.cl.Call(rpcIdGetFreshNum, make([]byte, 0), &reply, 100)
	if err == urpc.ErrNone {
		return DecodeUint64(reply), err
	}
	return 0, err
}

func (cl *Client) putRpc(args put_gk.S) Error {
	var reply []byte
	err := cl.cl.Call(rpcIdPut, put_gk.Marshal(args, make([]byte, 0)), &reply, 100)
	if err == urpc.ErrNone {
		return err
	}
	return err
}

func (cl *Client) conditionalPutRpc(args conditionalput_gk.S) (string, Error) {
	var reply []byte
	err := cl.cl.Call(rpcIdConditionalPut, conditionalput_gk.Marshal(args, make([]byte, 0)), &reply, 100)
	if err == urpc.ErrNone {
		return string(reply), err
	}
	return "", err
}

func (cl *Client) getRpc(args get_gk.S) (string, Error) {
	var reply []byte
	err := cl.cl.Call(rpcIdGet, get_gk.Marshal(args, make([]byte, 0)), &reply, 100)
	if err == urpc.ErrNone {
		return string(reply), err
	}
	return "", err
}

func makeClient(hostname grove_ffi.Address) *Client {
	return &Client{cl: urpc.MakeClient(hostname)}
}
