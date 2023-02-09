package main

import (
	"fmt"
	"time"

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
			fmt.Print("Txn 1 done\n")
			return true
		})
	}()

	time.Sleep(2*time.Second)
	func() {
		txnCk := txn.Begin(txnMgrHost, txnCoordHost, indexHost)
		txnCk.DoTxn(func(t *txn.Clerk) bool {
			if len(t.Get(0)) > 0 {
				machine.Assert(len(t.Get(1)) > 0)
				fmt.Print("Val: ", t.Get(1), "\n")
			}
			fmt.Print("Val: ", t.Get(1), "\n")
			return true
		})
	}()
}
