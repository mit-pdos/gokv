package lockservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

func (s *Server) Start(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[rpcIdGetFreshNum] =
		func(enc_args []byte, enc_reply *[]byte) {
			*enc_reply = EncodeUint64(s.getFreshNum())
		}

	handlers[rpcIdPut] =
		func(enc_args []byte, enc_reply *[]byte) {
			s.put(decodePutArgs(enc_args))
		}

	handlers[rpcIdConditionalPut] =
		func(enc_args []byte, enc_reply *[]byte) {
			*enc_reply = []byte(s.conditionalPut(decodeConditionalPutArgs(enc_args)))
		}

	handlers[rpcIdGet] =
		func(enc_args []byte, enc_reply *[]byte) {
			*enc_reply = []byte(s.get(decodeGetArgs(enc_args)))
		}

	urpc.MakeServer(handlers).Serve(me)
}
