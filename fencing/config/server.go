package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
	"sync"
)

const (
	TIMEOUT_MS = uint64(1000)
	MILLION    = uint64(1000000)
)

type Server struct {
	mu           *sync.Mutex
	data         grove_ffi.Address
	currentEpoch uint64
	epoch_cond   *sync.Cond

	currHolderActive      bool
	currHolderActive_cond *sync.Cond
	heartbeatExpiration   grove_ffi.Time
}

func (s *Server) AcquireEpoch(newFrontend grove_ffi.Address) uint64 {
	s.mu.Lock()
	for s.currHolderActive {
		s.currHolderActive_cond.Wait()
	}
	s.currHolderActive = true
	s.data = newFrontend
	s.currentEpoch += 1

	now := grove_ffi.TimeNow()
	s.heartbeatExpiration = now + TIMEOUT_MS*MILLION

	ret := s.currentEpoch

	s.mu.Unlock()
	return ret
}

func (s *Server) HeartbeatListener() {
	// sets currHolderActive to false, at most once per epoch.
	var epochToWaitFor uint64 = 1
	for {
		// If the current epoch expired and there's no one trying to get
		// ownership, we should just wait for the epoch to increase.
		// Alternatively, could kill the HeartbeatListener thread and start it
		// up when another client gets ownership
		s.mu.Lock()
		for s.currentEpoch < epochToWaitFor {
			s.epoch_cond.Wait()
		}
		s.mu.Unlock()

		// loops until the heartbeat expiration time is passed
		for {
			now := grove_ffi.TimeNow()
			s.mu.Lock()
			if now < s.heartbeatExpiration {
				delay := s.heartbeatExpiration - now
				s.mu.Unlock()
				grove_ffi.Sleep(delay)
			} else {
				s.currHolderActive = false
				s.currHolderActive_cond.Signal()
				epochToWaitFor = s.currentEpoch + 1
				s.mu.Unlock()
				break
			}
		}

	}
}

func (s *Server) Heartbeat(epoch uint64) bool {
	// reset the heartbeat expiration time
	s.mu.Lock()
	var ret bool = false
	if s.currentEpoch == epoch {
		now := grove_ffi.TimeNow()
		s.heartbeatExpiration = now + TIMEOUT_MS
		ret = true
	}
	s.mu.Unlock()
	return ret
}

// XXX: don't need to send fencing token here, because client won't need it
func (s *Server) Get() grove_ffi.Address {
	s.mu.Lock()
	ret := s.data
	s.mu.Unlock()
	return ret
}

func StartServer(me grove_ffi.Address) {
	s := new(Server)
	s.mu = new(sync.Mutex)
	s.data = 0
	s.currentEpoch = 0
	s.epoch_cond = sync.NewCond(s.mu)
	s.currHolderActive = false
	s.currHolderActive_cond = sync.NewCond(s.mu)

	go func() {
		s.HeartbeatListener()
	}()

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_LOCK] = func(args []byte, reply *[]byte) {
		dec := marshal.NewDec(args)
		enc := marshal.NewEnc(8)
		enc.PutInt(s.AcquireEpoch(dec.GetInt()))
		*reply = enc.Finish()
	}

	handlers[RPC_GET] = func(args []byte, reply *[]byte) {
		enc := marshal.NewEnc(8)
		enc.PutInt(s.Get())
		*reply = enc.Finish()
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
}
