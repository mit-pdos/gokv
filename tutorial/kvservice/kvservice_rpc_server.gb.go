package kvservice

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/kvservice/conditionalput_gk"
	"github.com/mit-pdos/gokv/tutorial/kvservice/get_gk"
	"github.com/mit-pdos/gokv/tutorial/kvservice/put_gk"
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
			args, _ := put_gk.Unmarshal(enc_args)
			s.put(args)
		}

	handlers[rpcIdConditionalPut] =
		func(enc_args []byte, enc_reply *[]byte) {
			args, _ := conditionalput_gk.Unmarshal(enc_args)
			*enc_reply = []byte(s.conditionalPut(args))
		}

	handlers[rpcIdGet] =
		func(enc_args []byte, enc_reply *[]byte) {
			args, _ := get_gk.Unmarshal(enc_args)
			*enc_reply = []byte(s.get(args))
		}

	urpc.MakeServer(handlers).Serve(me)
}
