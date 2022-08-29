package config

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type Server struct {
	mu     *sync.Mutex
	epoch  uint64
	config []grove_ffi.Address
}

func (s *Server) GetEpochAndConfig(args []byte, reply *[]byte) {
	s.mu.Lock()

	s.epoch = std.SumAssumeNoOverflow(s.epoch, 1)

	*reply = make([]byte, 0, 8+8*len(s.config))
	*reply = marshal.WriteInt(*reply, s.epoch)
	*reply = marshal.WriteBytes(*reply, EncodeConfig(s.config))
	s.mu.Unlock()
}

func (s *Server) GetConfig(args []byte, reply *[]byte) {
	s.mu.Lock()
	*reply = EncodeConfig(s.config)
	s.mu.Unlock()
}

func (s *Server) WriteConfig(args []byte, reply *[]byte) {
	s.mu.Lock()
	epoch, enc := marshal.ReadInt(args)
	if epoch < s.epoch {
		*reply = marshal.WriteInt(*reply, e.Stale)
		s.mu.Unlock()
		log.Println("Stale write", s.config)
		return
	}
	s.config = DecodeConfig(enc)
	log.Println("New config is:", s.config)
	*reply = marshal.WriteInt(*reply, e.None)
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
	handlers[RPC_GETCONFIG] = s.GetConfig
	handlers[RPC_WRITECONFIG] = s.WriteConfig
	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
