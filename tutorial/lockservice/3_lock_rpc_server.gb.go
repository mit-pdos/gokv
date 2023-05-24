package lockservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

func (s *Server) Start(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[RPC_GET_FRESH_NUM] =
		func(enc_args []byte, enc_reply *[]byte) {
			*enc_reply = EncodeUint64(s.getFreshNum())
		}

	handlers[RPC_TRY_ACQUIRE] =
		func(enc_args []byte, enc_reply *[]byte) {
			*enc_reply = EncodeUint64(s.tryAcquire(DecodeUint64(enc_args)))
		}

	handlers[RPC_RELEASE] =
		func(enc_args []byte, enc_reply *[]byte) {
			s.release(DecodeUint64(enc_args))
		}

	urpc.MakeServer(handlers).Serve(me)
}
