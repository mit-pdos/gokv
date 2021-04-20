package main

import (
	"fmt"
	"github.com/upamanyus/urpc/rpc"
	"testing"
	"time"
	"sync"
)

func TestRPC(t *testing.T) {
	fmt.Println("==Basic urpc test")
	client := rpc.MakeRPCClient("localhost:12345")

	var reply []byte
	args := make([]byte, 0)
	reply = make([]byte, 0)
	client.Call(1, args, &reply)
	fmt.Printf("%s\n", string(reply))
}

func TestBenchRPC(t *testing.T) {
	fmt.Println("==Benchmarking urpc")
	client := rpc.MakeRPCClient("localhost:12345")

	start := time.Now()
	N := 200000
	var reply []byte
	args := make([]byte, 0)
	for n := 0; n < N; n++ {
		reply = make([]byte, 0)
		client.Call(1, args, &reply)
	}
	d := time.Since(start)
	fmt.Printf("%v us/op\n", d.Microseconds()/int64(N))
	fmt.Printf("%v ops/sec\n", float64(N) / d.Seconds())
}

func TestBenchConcurrentRPC(t *testing.T) {
	fmt.Println("==Benchmarking urpc")
	numClients := 40
	clients := make([]*rpc.RPCClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = rpc.MakeRPCClient("localhost:12345")
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
				clients[i].Call(1, args, &reply)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	d := time.Since(start)
	fmt.Printf("%v us/op\n", d.Microseconds()/int64(N))
	fmt.Printf("%v ops/sec\n", float64(N * numClients) / d.Seconds())
}
