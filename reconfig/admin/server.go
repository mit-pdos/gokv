package admin_server

import (
	// "github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconfig/util"
	"sync"
)

type Server struct {
	mu       *sync.Mutex
	latestCn uint64
	// clerk    *config.Clerk
}

func (s *Server) ReconfigureTo(c *util.Configuration) {
	// newMarshalledConfig := util.EncodeConfiguration(c)
}
