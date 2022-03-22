package frontend

import (
	"github.com/mit-pdos/gokv/fencing/config"
	"github.com/mit-pdos/gokv/fencing/ctr"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu *sync.Mutex

	epoch uint64

	ck1  *ctr.Clerk
	ctr1 uint64

	ck2  *ctr.Clerk
	ctr2 uint64
}

// pre: key == 0 or key == 1
func (s *Server) FetchAndIncrement(key uint64) uint64 {
	s.mu.Lock()
	var ret uint64
	if key == 0 {
		s.ck1.Put(s.ctr1+1, s.epoch)
		s.ctr1 += 1
		ret = s.ctr1
	} else {
		// key == 1
		s.ck2.Put(s.ctr2+1, s.epoch)
		s.ctr2 += 1
		ret = s.ctr2
	}
	s.mu.Unlock()
	return ret
}

func StartServer(me, configHost, host1, host2 grove_ffi.Address) {
	s := new(Server)

	configCk := config.MakeClerk(configHost)
	s.epoch = configCk.Lock(me)

	s.mu = new(sync.Mutex)
	s.ck1 = ctr.MakeClerk(host1)
	s.ck2 = ctr.MakeClerk(host2)

	s.ctr1 = s.ck1.Get(s.epoch)
	s.ctr2 = s.ck2.Get(s.epoch)

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_FAI] = func(args []byte, reply *[]byte) {
		dec := marshal.NewDec(args)
		enc := marshal.NewEnc(8)
		enc.PutInt(s.FetchAndIncrement(dec.GetInt()))
		*reply = enc.Finish()
	}

	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)
}
