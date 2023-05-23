package txnmgr

import (
	"github.com/tchajed/goose/machine"
	"sync"
)

type Server struct {
	mu      *sync.Mutex
	p       machine.ProphId
	nextTid uint64
}

func MakeServer() *Server {
	p := machine.NewProph()
	txnMgr := &Server{p: p, nextTid: 1}
	txnMgr.mu = new(sync.Mutex)
	return txnMgr
}

func (txnMgr *Server) New() uint64 {
	txnMgr.mu.Lock()
	tid := txnMgr.nextTid
	txnMgr.nextTid += 1
	txnMgr.mu.Unlock()
	return tid
}
