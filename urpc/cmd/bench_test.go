package main

import (
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"sync"
	"testing"
	"time"
)

func TestRPC(t *testing.T) {
	fmt.Println("==Basic urpc test")
	client := urpc.MakeClient(grove_ffi.MakeAddress("127.0.0.1:12345"))

	var reply []byte
	args := make([]byte, 0)
	reply = make([]byte, 0)
	client.Call(1, args, &reply, 1000 /* ms */)
	fmt.Printf("%s\n", string(reply))
}

func TestBenchRPC(t *testing.T) {
	fmt.Println("==Benchmarking urpc")
	client := urpc.MakeClient(grove_ffi.MakeAddress("127.0.0.1:12345"))

	start := time.Now()
	N := 200000
	var reply []byte
	args := make([]byte, 0)
	for n := 0; n < N; n++ {
		reply = make([]byte, 0)
		client.Call(1, args, &reply, 1000 /* ms */)
	}
	d := time.Since(start)
	fmt.Printf("%v us/op\n", d.Microseconds()/int64(N))
	fmt.Printf("%v ops/sec\n", float64(N)/d.Seconds())
}

func TestBenchConcurrentRPC(t *testing.T) {
	fmt.Println("==Benchmarking urpc")
	numClients := 40
	clients := make([]*urpc.Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = urpc.MakeClient(grove_ffi.MakeAddress("127.0.0.1:12345"))
	}

	var wg sync.WaitGroup
	N := 50000
	start := time.Now()
	args := make([]byte, 128)
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(i int) {
			var reply []byte
			for n := 0; n < N; n++ {
				reply = make([]byte, 0)
				clients[i].Call(1, args, &reply, 1000 /* ms */)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	d := time.Since(start)
	fmt.Printf("%v us/op\n", d.Microseconds()/int64(N))
	fmt.Printf("%v ops/sec\n", float64(N*numClients)/d.Seconds())
}
