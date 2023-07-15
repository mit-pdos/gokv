package txn

import (
	"github.com/mit-pdos/vmvcc/trusted_proph"
	"github.com/mit-pdos/gokv/dmvcc/index"
	"github.com/mit-pdos/gokv/dmvcc/prophname"
	"github.com/mit-pdos/gokv/dmvcc/txncoordinator"
	"github.com/mit-pdos/gokv/dmvcc/txnmgr"

	// "github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/goose/machine"
)

type Clerk struct {
	p          machine.ProphId
	tid        uint64
	writes     map[uint64]string
	indexCk    *index.Clerk
	txnMgrHost *txnmgr.Server
	txnCoordCk *txncoordinator.Clerk
}

func Begin(txnMgrHost *txnmgr.Server, txnCoordHost *txncoordinator.Server,
	indexHost *index.Server) *Clerk {
	return &Clerk{
		p:          prophname.Get(),
		tid:        txnMgrHost.New(), // cheating a bit here
		writes:     make(map[uint64]string),
		indexCk:    index.MakeClerk(indexHost),
		txnMgrHost: txnMgrHost,
		txnCoordCk: txncoordinator.MakeClerk(txnCoordHost),
	}
}

func (txnCk *Clerk) Put(key uint64, val string) {
	txnCk.writes[key] = val
}

func (txnCk *Clerk) Get(key uint64) string {
	val1, ok := txnCk.writes[key]
	if ok {
		return val1
	}

	val2 := txnCk.indexCk.Read(key, txnCk.tid)
	trusted_proph.ResolveRead(txnCk.p, txnCk.tid, key)
	return val2
}

func (txnCk *Clerk) abort() {
	trusted_proph.ResolveAbort(txnCk.p, txnCk.tid)
}

func (txn *Clerk) begin() {
	tid := txn.txnMgrHost.New()
	txn.tid = tid
}

func (txn *Clerk) DoTxn(body func(txn *Clerk) bool) bool {
	txn.begin()

	wantToCommit := body(txn)
	if !wantToCommit {
		txn.abort()
		return false
	}

	return txn.txnCoordCk.TryCommit(txn.tid, txn.writes)
}
