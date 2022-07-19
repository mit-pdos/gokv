package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	// "github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu     *sync.Mutex
	epoch  uint64
	config []grove_ffi.Address
}

func (s *Server) GetEpochAndConfig(args []byte, reply *[]byte) {

}

func (s *Server) WriteConfig(args []byte, reply *[]byte) {
	s.mu.Lock()
	// epoch, enc := marshal.ReadInt(args)
	// if epoch < s.epoch {
	// s.mu.Unlock()
	// return
	// }

}
