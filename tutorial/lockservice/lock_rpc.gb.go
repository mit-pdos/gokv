package lockservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

const (
	RPC_GET_FRESH_NUM = uint64(0)
	RPC_TRY_ACQUIRE = uint64(1)
	RPC_RELEASE = uint64(2)
)

type Error = uint64

func (s *Server) Start(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte)[]byte)

	handlers[RPC_GET_FRESH_NUM] =
		func(enc_args []byte) []byte {
			return EncodeUint64(s.getFreshNum())
		}

	handlers[RPC_TRY_ACQUIRE] =
		func(enc_args []byte) []byte {
			return EncodeBool(s.tryAcquire(DecodeUint64(enc_args)))
		}

	handlers[RPC_RELEASE] =
		func(enc_args []byte) []byte {
			s.release(DecodeUint64(enc_args))
			return nil
		}

	urpc.MakeServer(handlers).Serve(me)
}

type Client struct {
	cl *urpc.Client
}

func (cl *Client) getFreshNum() (uint64, Error) {
	args := make([]byte, 0)
	reply, err := cl.cl.Call(RPC_GET_FRESH_NUM, args, 100)
	if err == urpc.ErrNone {
		return DecodeUint64(reply), err
	}
	return 0, err
}

func (cl *Client) tryAcquire(id uint64) (bool, Error) {
	args := EncodeUint64(id)
	reply, err := cl.cl.Call(RPC_GET_FRESH_NUM, args, 100)
	if err == urpc.ErrNone {
		return DecodeBool(reply), err
	}
	return false, err
}

func (cl *Client) release(id uint64) Error {
	args := EncodeUint64(id)
	_, err := cl.cl.Call(RPC_RELEASE, args, 100)
	return err
}
