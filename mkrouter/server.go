package mkrouter

import (
	"sync"

	"github.com/mit-pdos/gokv/memkv"
	"github.com/mit-pdos/gokv/memkv/getrequest_gk"
	"github.com/mit-pdos/gokv/memkv/kvop_gk"
	"github.com/mit-pdos/gokv/memkv/putrequest_gk"
)

// Totally naive RPC based MemKV proxy/load balancer

type HostName = uint64

type MKRouterServer struct {
	coord HostName
	mu    *sync.Mutex
	cks   []*memkv.KVClerk // pool of clerks
}

func (s *MKRouterServer) GetRPC(args *getrequest_gk.S, val *[]byte) {
	var ck *memkv.KVClerk
	s.mu.Lock()
	if len(s.cks) > 0 {
		ck = s.cks[0]
		s.mu.Unlock()
	} else {
		s.mu.Unlock()
		ck = memkv.MakeKVClerk(s.coord, nil) // FIXME: needs a connman
	}

	*val = ck.Get(args.Key)
	s.mu.Lock()
	s.cks = append(s.cks, ck)
	s.mu.Unlock()
}

func (s *MKRouterServer) PutRPC(args *putrequest_gk.S) {
	var ck *memkv.KVClerk
	s.mu.Lock()
	if len(s.cks) > 0 {
		ck = s.cks[0]
		s.mu.Unlock()
	} else {
		s.mu.Unlock()
		ck = memkv.MakeKVClerk(s.coord, nil) // FIXME: needs a connman
	}

	ck.Put(args.Key, args.Value)
	s.mu.Lock()
	s.cks = append(s.cks, ck)
	s.mu.Unlock()
}

func (mkv *MKRouterServer) Start(host HostName) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[uint64(kvop_gk.KV_PUT)] = func(rawReq []byte, rawReply *[]byte) {
		req, _ := putrequest_gk.Unmarshal(rawReq)
		mkv.PutRPC(&req)
	}

	handlers[uint64(kvop_gk.KV_GET)] = func(rawReq []byte, rawReply *[]byte) {
		req, _ := getrequest_gk.Unmarshal(rawReq)
		mkv.GetRPC(&req, rawReply)
	}
}
