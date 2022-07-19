package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu     *sync.Mutex
	epoch  uint64
	config []grove_ffi.Address
}

func (s *Server) GetEpochAndConfig(args []byte, reply *[]byte) {
	s.mu.Lock()
	s.epoch += 1
	*reply = make([]byte, 0, 8+8*len(s.config))
	*reply = marshal.WriteInt(*reply, s.epoch)
	*reply = marshal.WriteBytes(*reply, EncodeConfig(s.config))
	s.mu.Unlock()
}

func (s *Server) WriteConfig(args []byte, reply *[]byte) {
	s.mu.Lock()
	epoch, enc := marshal.ReadInt(args)
	if epoch < s.epoch {
		s.mu.Unlock()
		return
	}
	s.config = DecodeConfig(enc)
	s.mu.Unlock()
}

func MakeServer() *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.epoch = 0
	s.config = make([]grove_ffi.Address, 0)
	return s
}

func (s *Server) Serve(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_GETEPOCH] = s.GetEpochAndConfig
	handlers[RPC_WRITECONFIG] = s.WriteConfig
	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
