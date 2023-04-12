package config

import (
	"log"
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
)

const LeaseInterval = uint64(1_000_000_000) // 1 second

type Server struct {
	mu                *sync.Mutex
	epoch             uint64
	leaseExpiration   uint64
	wantLeaseToExpire bool
	config            []grove_ffi.Address
}

func (s *Server) GetEpochAndConfig(args []byte, reply *[]byte) {
	// check if lease is expired
	for {
		l, _ := grove_ffi.GetTimeRange()
		s.mu.Lock()
		if l >= s.leaseExpiration {
			break
		} else {
			s.wantLeaseToExpire = true
			s.mu.Unlock()
			machine.Sleep(s.leaseExpiration - l) // sleep long enough for lease to be expired
		}
	}
	s.wantLeaseToExpire = false

	s.epoch = std.SumAssumeNoOverflow(s.epoch, 1)
	*reply = make([]byte, 0, 8+8+8*len(s.config))
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
	if epoch != s.epoch {
		*reply = marshal.WriteInt(nil, e.Stale)
		s.mu.Unlock()
		log.Println("Stale write", s.config)
		return
	}
	s.config = DecodeConfig(enc)
	log.Println("New config is:", s.config)
	*reply = marshal.WriteInt(nil, e.None)
	s.mu.Unlock()
}

func (s *Server) GetLease(args []byte, reply *[]byte) {
	epoch, _ := marshal.ReadInt(args)
	s.mu.Lock()
	if s.epoch != epoch || s.wantLeaseToExpire {
		s.mu.Unlock()
		*reply = marshal.WriteInt(nil, e.Stale)
		*reply = marshal.WriteInt(*reply, 0)
		log.Println("Rejected lease request", epoch, s.epoch, s.wantLeaseToExpire)
		return
	}

	l, _ := grove_ffi.GetTimeRange()
	newLeaseExpiration := l + LeaseInterval
	if newLeaseExpiration > s.leaseExpiration {
		s.leaseExpiration = newLeaseExpiration
	}
	s.mu.Unlock()

	*reply = marshal.WriteInt(nil, e.None)
	*reply = marshal.WriteInt(*reply, newLeaseExpiration)
}

func MakeServer(initconfig []grove_ffi.Address) *Server {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.epoch = 0
	s.config = initconfig
	return s
}

func (s *Server) Serve(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[RPC_GETEPOCH] = s.GetEpochAndConfig
	handlers[RPC_GETCONFIG] = s.GetConfig
	handlers[RPC_WRITECONFIG] = s.WriteConfig
	handlers[RPC_GETLEASE] = s.GetLease

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
