package config2

import (
	"log"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/mpaxos"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
)

const LeaseInterval = uint64(1_000_000_000) // 1 second

type state struct {
	epoch             uint64
	reservedEpoch     uint64
	leaseExpiration   uint64
	wantLeaseToExpire bool
	config            []grove_ffi.Address
}

func encodeState(st *state) []byte {
	var e []byte
	e = marshal.WriteInt(nil, st.epoch)
	e = marshal.WriteInt(e, st.reservedEpoch)
	e = marshal.WriteInt(e, st.leaseExpiration)
	if st.wantLeaseToExpire {
		e = marshal.WriteInt(e, 1)
	} else {
		e = marshal.WriteInt(e, 0)
	}
	e = marshal.WriteBytes(e, EncodeConfig(st.config))
	return e
}

func decodeState(e []byte) *state {
	st := new(state)
	var e2 []byte = e
	st.epoch, e2 = marshal.ReadInt(e2)
	st.reservedEpoch, e2 = marshal.ReadInt(e2)
	st.leaseExpiration, e2 = marshal.ReadInt(e2)
	var wantExp uint64
	wantExp, e2 = marshal.ReadInt(e2)
	if wantExp != 0 {
		st.wantLeaseToExpire = true
	}
	st.config = DecodeConfig(e2)
	return st
}

type Server struct {
	s *mpaxos.Server
}

// TODO: mpaxos doesn't need to return reply anymore
func (s *Server) withLock(f func(st *state)) {
	s.s.Apply(func(e []byte) ([]byte, []byte) {
		st := decodeState(e)
		f(st)
		return encodeState(st), nil
	})
}

func (s *Server) ReserveEpochAndGetConfig(args []byte, reply *[]byte) {
	s.withLock(func(st *state) {
		st.reservedEpoch = std.SumAssumeNoOverflow(st.reservedEpoch, 1)
		*reply = make([]byte, 0, 8+8+8*len(st.config))
		*reply = marshal.WriteInt(*reply, st.reservedEpoch)
		*reply = marshal.WriteBytes(*reply, EncodeConfig(st.config))
	})
}

func (s *Server) GetConfig(args []byte, reply *[]byte) {
	st := decodeState(s.s.WeakRead())
	*reply = EncodeConfig(st.config)
}

func (s *Server) TryWriteConfig(args []byte, reply *[]byte) {
	// check if lease is expired
	epoch, enc := marshal.ReadInt(args)
	for {
		var done bool = false
		var timeToSleep uint64
		s.withLock(func(st *state) {
			if epoch < st.reservedEpoch {
				*reply = marshal.WriteInt(nil, e.Stale)
				log.Printf("Stale: %d < %d", epoch, st.reservedEpoch)
				done = true
				return
			} else if epoch > st.epoch {
				l, _ := grove_ffi.GetTimeRange()
				if l >= st.leaseExpiration {
					st.wantLeaseToExpire = false
					st.epoch = epoch

					st.config = DecodeConfig(enc)
					log.Println("New config is:", st.config)
					*reply = marshal.WriteInt(nil, e.None)
					done = true
					return
				} else {
					st.wantLeaseToExpire = true
					timeToSleep = st.leaseExpiration - l
					done = false
					return
				}
			} else {
				// already in the epoch
				st.config = DecodeConfig(enc)
				// TODO: avoid putting marshalling in the critical section
				// s.mu.Unlock()
				*reply = marshal.WriteInt(nil, e.None)
				done = true
				return
			}
		})
		if done {
			break
		}
		machine.Sleep(timeToSleep) // sleep long enough for lease to be expired
		continue
	}
}

func (s *Server) GetLease(args []byte, reply *[]byte) {
	epoch, _ := marshal.ReadInt(args)
	var newLeaseExpiration uint64
	s.withLock(func(st *state) {
		if st.epoch != epoch || st.wantLeaseToExpire {
			// s.mu.Unlock()
			*reply = marshal.WriteInt(nil, e.Stale)
			*reply = marshal.WriteInt(*reply, 0)
			log.Println("Rejected lease request", epoch, st.epoch, st.wantLeaseToExpire)
			return
		}

		l, _ := grove_ffi.GetTimeRange()
		newLeaseExpiration := l + LeaseInterval
		if newLeaseExpiration > st.leaseExpiration {
			st.leaseExpiration = newLeaseExpiration
		}
	})

	if len(*reply) == 0 {
		*reply = marshal.WriteInt(nil, e.None)
		*reply = marshal.WriteInt(*reply, newLeaseExpiration)
	}
}

func MakeServer(initconfig []grove_ffi.Address) *Server {
	s := new(Server)
	// TODO: init
	// s.s = new(sync.Mutex)
	return s
}

func (s *Server) Serve(me grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[RPC_RESERVEEPOCH] = s.ReserveEpochAndGetConfig
	handlers[RPC_GETCONFIG] = s.GetConfig
	handlers[RPC_TRYWRITECONFIG] = s.TryWriteConfig
	handlers[RPC_GETLEASE] = s.GetLease

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
}
