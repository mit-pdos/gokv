package main

import (
	"fmt"

	"github.com/mit-pdos/gokv/dmvcc/index"
	"github.com/mit-pdos/gokv/dmvcc/txn"
	"github.com/mit-pdos/gokv/dmvcc/txncoordinator"
	"github.com/mit-pdos/gokv/dmvcc/txnmgr"
	"github.com/tchajed/goose/machine"
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
				machine.Assert(len(t.Get(1)) > 0)
			}
			fmt.Print("Val on txn2: \"", t.Get(1), "\"\n")
			return true
		})
	}()

	machine.Sleep(1e8)
}
