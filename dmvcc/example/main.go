package main

import (
	"log"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/dmvcc/index"
	"github.com/mit-pdos/gokv/dmvcc/txn"
	"github.com/mit-pdos/gokv/dmvcc/txncoordinator"
	"github.com/mit-pdos/gokv/dmvcc/txnmgr"
)

func main() {
	indexHost := index.MakeServer()
	txnMgrHost := txnmgr.MakeServer()
	txnCoordHost := txncoordinator.MakeServer(indexHost)

	go func() {
		txnCk := txn.Begin(txnMgrHost, txnCoordHost, indexHost)
		txnCk.DoTxn(func(t *txn.Clerk) bool {
			t.Put(0, "hello")
			t.Put(1, "world")
			return true
		})
	}()

	go func() {
		txnCk := txn.Begin(txnMgrHost, txnCoordHost, indexHost)
		txnCk.DoTxn(func(t *txn.Clerk) bool {
			if len(t.Get(0)) > 0 {
				primitive.Assert(len(t.Get(1)) > 0)
			}
			log.Printf("Val on txn2: '%s'", t.Get(1))
			return true
		})
	}()

	primitive.Sleep(uint64(100_000_000))
}
