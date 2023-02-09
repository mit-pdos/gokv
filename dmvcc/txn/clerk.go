package txn

import (

	"github.com/mit-pdos/go-mvcc/trusted_proph"
	"github.com/mit-pdos/gokv/dmvcc/index"
	"github.com/mit-pdos/gokv/dmvcc/prophname"
	"github.com/mit-pdos/gokv/dmvcc/txncoordinator"
	"github.com/mit-pdos/gokv/dmvcc/txnmgr"

	// "github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/goose/machine"
)

type Clerk struct {
	p       machine.ProphId
	tid     uint64
	writes  map[uint64]string
	indexCk *index.Clerk
	// txnCoordHost grove_ffi.Address
	txnCoordHost *txncoordinator.Server
}

func Begin(txnMgrHost *txnmgr.Server, txnCoordHost *txncoordinator.Server,
	indexHost *index.Server) *Clerk {
	return &Clerk{
		p:            prophname.Get(),
		tid:          txnMgrHost.New(), // cheating a bit here
		writes:       make(map[uint64]string),
		indexCk:      index.MakeClerk(indexHost),
		txnCoordHost: txnCoordHost,
	}
}

func (txnCk *Clerk) Put(key uint64, val string) {
	txnCk.writes[key] = val
}

func (txnCk *Clerk) Get(key uint64) string {
	val, ok := txnCk.writes[key]
	if ok {
		return val
	}

	val = txnCk.indexCk.Read(key, txnCk.tid)
	trusted_proph.ResolveRead(txnCk.p, txnCk.tid, key)
	return val
}

func (txnCk *Clerk) abort() {
	trusted_proph.ResolveAbort(txnCk.p, txnCk.tid)
}

func (txn *Clerk) DoTxn(body func(txn *Clerk) bool) bool {
	wantToCommit := body(txn)
	if !wantToCommit {
		txn.abort()
		return false
	}

	tcCk := txncoordinator.MakeClerk(txn.txnCoordHost)
	return tcCk.TryCommit(txn.tid, txn.writes)
}
