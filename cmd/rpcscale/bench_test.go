package main

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"math/rand"
	_ "net/http/pprof"
	"testing"
)

func BenchmarkNullRPC(b *testing.B) {
	cl := rpc.MakeRPCClient(grove_ffi.MakeAddress("0.0.0.0:12345"))
	rand.Seed(int64(b.N))
	reply := new([]byte)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cl.Call(RPC_NULL, nil, reply)
	}
}

func BenchmarkConcurrentNullRPC(b *testing.B) {
	rand.Seed(int64(b.N))

	b.ResetTimer()
	b.RunParallel(
		func(pb *testing.PB) {
			cl := rpc.MakeRPCClient(grove_ffi.MakeAddress("0.0.0.0:12345"))
			reply := new([]byte)
			for pb.Next() {
				cl.Call(RPC_NULL, nil, reply)
			}
		})
}
