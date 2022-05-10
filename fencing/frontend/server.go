package frontend

import (
	"github.com/mit-pdos/gokv/fencing/config"
	"github.com/mit-pdos/gokv/fencing/ctr"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
	"sync"
)

type Server struct {
	mu *sync.Mutex

	epoch uint64

	ck1 *ctr.Clerk
	ck2 *ctr.Clerk
}

// pre: key == 0 or key == 1
func (s *Server) FetchAndIncrement(key uint64) uint64 {
	s.mu.Lock()
	var ret uint64
	if key == 0 {
		ret = s.ck1.Get(s.epoch)
		s.ck1.Put(ret+1, s.epoch)
	} else {
		// key == 1
		ret = s.ck2.Get(s.epoch)
		s.ck2.Put(ret+1, s.epoch)
	}
	s.mu.Unlock()
	return ret
}

func StartServer(me, configHost, host1, host2 grove_ffi.Address) {
	s := new(Server)

	configCk := config.MakeClerk(configHost)
	s.epoch = configCk.AcquireEpoch(me)

	s.mu = new(sync.Mutex)
	s.ck1 = ctr.MakeClerk(host1)
	s.ck2 = ctr.MakeClerk(host2)

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RPC_FAI] = func(args []byte, reply *[]byte) {
		dec := marshal.NewDec(args)
		enc := marshal.NewEnc(8)
		enc.PutInt(s.FetchAndIncrement(dec.GetInt()))
		*reply = enc.Finish()
	}

	r := urpc.MakeServer(handlers)
	r.Serve(me)
}
