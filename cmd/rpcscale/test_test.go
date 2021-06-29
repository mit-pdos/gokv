package main

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	_ "net/http/pprof"
	"sync/atomic"
	"testing"
	"time"
)

func benchConcurrentNullRPC(numClients int) {
	reply := new([]byte)
	totalOps := new(uint64)

	// reporting thread
	go func() {
		p := message.NewPrinter(language.English)
		reportInterval := 1 * time.Second
		for {
			prevOps := atomic.LoadUint64(totalOps)
			time.Sleep(reportInterval)
			currOps := atomic.LoadUint64(totalOps) // BUG: should check the amount of elapsed time to be more accurate, rather than assuming it's just reportInterval exactly
			p.Printf("%d ops/sec in the past %v\n", currOps-prevOps, reportInterval)
		}

	}()

	for i := 0; i < numClients; i++ {
		cl := rpc.MakeRPCClient(grove_ffi.MakeAddress("0.0.0.0:12345"))
		go func() {
			for {
				cl.Call(RPC_NULL, nil, reply, 100)
				atomic.AddUint64(totalOps, 1)
			}
		}()
	}
	select {}
}

func TestBenchConcurrentRPC(t *testing.T) {
	benchConcurrentNullRPC(8)
}
