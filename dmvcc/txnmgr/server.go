package txnmgr

import (
	"github.com/goose-lang/primitive"
	"sync"
)

type Server struct {
	mu      *sync.Mutex
	p       primitive.ProphId
	nextTid uint64
}

func MakeServer() *Server {
	p := primitive.NewProph()
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
