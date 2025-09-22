package configservice

import (
	"log"

	"github.com/goose-lang/primitive"
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/mit-pdos/gokv/vrsm/configservice/config_gk"
	"github.com/mit-pdos/gokv/vrsm/paxos"
	"github.com/mit-pdos/gokv/vrsm/replica/err_gk"
	"github.com/tchajed/marshal"
)

const LeaseInterval = uint64(1_000_000_000) // 1 second

type state struct {
	epoch             uint64
	reservedEpoch     uint64
	leaseExpiration   uint64
	wantLeaseToExpire bool
	config            config_gk.S
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
	e = config_gk.Marshal(e, st.config)
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
	st.wantLeaseToExpire = (wantExp == 1)
	st.config, _ = config_gk.Unmarshal(e2)
	return st
}

type Server struct {
	s *paxos.Server
}

func (s *Server) tryAcquire() (bool, *state, func() bool) {
	err, e, relF := s.s.TryAcquire()
	if err != 0 {
		var p *state // XXX: hack to return nil pointer for goose
		return false, p, nil
	}
	st := decodeState(*e)
	releaseFn := func() bool {
		*e = encodeState(st)
		return (relF() == 0)
	}
	return true, st, releaseFn
}

func (s *Server) ReserveEpochAndGetConfig(args []byte, reply *[]byte) {
	*reply = err_gk.Marshal(make([]byte, 0), err_gk.NotLeader)
	ok, st, tryReleaseFn := s.tryAcquire()
	if !ok {
		return
	}
	st.reservedEpoch = std.SumAssumeNoOverflow(st.reservedEpoch, 1)
	config := st.config
	reservedEpoch := st.reservedEpoch
	if !tryReleaseFn() {
		return
	}
	*reply = make([]byte, 0, 8+8+8*len(config.Addrs))
	*reply = err_gk.Marshal(*reply, err_gk.None)
	*reply = marshal.WriteInt(*reply, reservedEpoch)
	*reply = config_gk.Marshal(*reply, config)
}

func (s *Server) GetConfig(args []byte, reply *[]byte) {
	st := decodeState(s.s.WeakRead())
	*reply = config_gk.Marshal(make([]byte, 0), st.config)
}

func (s *Server) TryWriteConfig(args []byte, reply *[]byte) {
	*reply = err_gk.Marshal(make([]byte, 0), err_gk.NotLeader)

	// check if lease is expired
	epoch, enc := marshal.ReadInt(args)
	config, _ := config_gk.Unmarshal(enc)
	for {
		ok, st, tryReleaseFn := s.tryAcquire()
		if !ok {
			break
		}

		if epoch < st.reservedEpoch {
			if !tryReleaseFn() {
				break
			}
			*reply = err_gk.Marshal(make([]byte, 0), err_gk.Stale)
			log.Printf("Stale: %d < %d", epoch, st.reservedEpoch)
			break
		} else if epoch > st.epoch {
			l, _ := grove_ffi.GetTimeRange()
			if l >= st.leaseExpiration {
				st.wantLeaseToExpire = false
				st.epoch = epoch
				st.config = config
				if !tryReleaseFn() {
					break
				}
				log.Println("New config is:", st.config)
				*reply = err_gk.Marshal(make([]byte, 0), err_gk.None)
				break
			} else {
				st.wantLeaseToExpire = true
				timeToSleep := st.leaseExpiration - l
				if !tryReleaseFn() {
					break
				}
				primitive.Sleep(timeToSleep) // sleep long enough for lease to be expired
				continue
			}
		} else {
			// already in the epoch
			st.config = config
			if !tryReleaseFn() {
				break
			}
			*reply = err_gk.Marshal(make([]byte, 0), err_gk.None)
			break
		}
	}
}

func (s *Server) GetLease(args []byte, reply *[]byte) {
	*reply = err_gk.Marshal(make([]byte, 0), err_gk.NotLeader)
	*reply = marshal.WriteInt(*reply, 0) // placeholder lease expiration time
	epoch, _ := marshal.ReadInt(args)
	ok, st, tryReleaseFn := s.tryAcquire()
	if !ok {
		return
	}

	if st.epoch != epoch || st.wantLeaseToExpire {
		log.Println("Rejected lease request", epoch, st.epoch, st.wantLeaseToExpire)
		if !tryReleaseFn() {
			return
		}
		*reply = err_gk.Marshal(make([]byte, 0), err_gk.Stale)
		*reply = marshal.WriteInt(*reply, 0)
		return
	}

	l, _ := grove_ffi.GetTimeRange()
	newLeaseExpiration := l + LeaseInterval
	if newLeaseExpiration > st.leaseExpiration {
		st.leaseExpiration = newLeaseExpiration
	}
	if !tryReleaseFn() {
		return
	}

	*reply = err_gk.Marshal(make([]byte, 0), err_gk.None)
	*reply = marshal.WriteInt(*reply, newLeaseExpiration)
}

func makeServer(fname string, paxosMe grove_ffi.Address,
	hosts []grove_ffi.Address, initconfig []grove_ffi.Address) *Server {
	s := new(Server)
	initEnc := encodeState(&state{config: config_gk.S{Addrs: initconfig}})

	s.s = paxos.StartServer(fname, initEnc, paxosMe, hosts)

	return s
}

func StartServer(fname string, me grove_ffi.Address, paxosMe grove_ffi.Address,
	hosts []grove_ffi.Address, initconfig []grove_ffi.Address) *Server {
	s := makeServer(fname, paxosMe, hosts, initconfig)
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[RPC_RESERVEEPOCH] = s.ReserveEpochAndGetConfig
	handlers[RPC_GETCONFIG] = s.GetConfig
	handlers[RPC_TRYWRITECONFIG] = s.TryWriteConfig
	handlers[RPC_GETLEASE] = s.GetLease

	rs := urpc.MakeServer(handlers)
	rs.Serve(me)
	return s
}
