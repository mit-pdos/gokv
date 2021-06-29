package mkrouter

import (
	"github.com/mit-pdos/gokv/memkv"
	"sync"
)

// Totally naive RPC based MemKV proxy/load balancer

type HostName = uint64

type MKRouterServer struct {
	coord HostName
	mu    *sync.Mutex
	cks   []*memkv.MemKVClerk // pool of clerks
}

func (s *MKRouterServer) GetRPC(args *memkv.GetRequest, val *[]byte) {
	var ck *memkv.MemKVClerk
	s.mu.Lock()
	if len(s.cks) > 0 {
		ck = s.cks[0]
		s.mu.Unlock()
	} else {
		s.mu.Unlock()
		ck = memkv.MakeMemKVClerk(s.coord)
	}

	*val = ck.Get(args.Key)
	s.mu.Lock()
	s.cks = append(s.cks, ck)
	s.mu.Unlock()
}

func (s *MKRouterServer) PutRPC(args *memkv.PutRequest) {
	var ck *memkv.MemKVClerk
	s.mu.Lock()
	if len(s.cks) > 0 {
		ck = s.cks[0]
		s.mu.Unlock()
	} else {
		s.mu.Unlock()
		ck = memkv.MakeMemKVClerk(s.coord)
	}

	ck.Put(args.Key, args.Value)
	s.mu.Lock()
	s.cks = append(s.cks, ck)
	s.mu.Unlock()
}

func (mkv *MKRouterServer) Start(host HostName) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[memkv.KV_PUT] = func(rawReq []byte, rawReply *[]byte) {
		mkv.PutRPC(memkv.DecodePutRequest(rawReq))
	}

	handlers[memkv.KV_GET] = func(rawReq []byte, rawReply *[]byte) {
		mkv.GetRPC(memkv.DecodeGetRequest(rawReq), rawReply)
	}
}
