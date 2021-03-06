package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/upamanyu/gokv"
	"strconv"
	"sync/atomic"
	"time"
)

func repeat_until_done(f func(int), done *int32) {
	go func() {
		for i := 0; atomic.LoadInt32(done) == 0; i++ {
			f(i)
		}
	}()
}

type PutThroughputExperiment struct {
	NumClients     int
	NumKeys        int
	WarmupTime     time.Duration
	ExperimentTime time.Duration
}

func (e *PutThroughputExperiment) run() {
	fmt.Printf("==Testing put throughput with %+v\n", *e)
	var ck []*gokv.GoKVClerk
	ck = make([]*gokv.GoKVClerk, e.NumClients)
	for i := 0; i < e.NumClients; i++ {
		ck[i] = gokv.MakeKVClerk(uint64(i), "127.0.0.1")
	}

	var done *int32 = new(int32)

	numOps := new(uint64)
	for i := 0; i < e.NumClients; i++ {
		ck1 := ck[i]
		repeat_until_done(func(j int) {
			ck1.Put(uint64(j%e.NumKeys), "somevalue")
			atomic.AddUint64(numOps, 1)
		}, done)
	}

	time.Sleep(e.WarmupTime)
	fmt.Println("Warmup done, starting experiment")
	atomic.StoreUint64(numOps, 0)
	time.Sleep(e.ExperimentTime)
	nOp := atomic.LoadUint64(numOps)
	atomic.StoreInt32(done, 1)
	fmt.Printf("%f puts/sec\n", float64(nOp)/e.ExperimentTime.Seconds())
}

type RedisPutThroughputExperiment struct {
	NumClients     int
	NumKeys        int
	WarmupTime     time.Duration
	ExperimentTime time.Duration
}

var ctx = context.Background()

func doRedisPut() {
	cl := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	err := cl.Set(ctx, "test", "test", 0).Err()
	if err != nil {
		panic(err)
	}
}

func (e *RedisPutThroughputExperiment) run() {
	fmt.Printf("==Testing redis put throughput with %+v\n", *e)
	doRedisPut() // make sure server is up

	var cks []*redis.Client
	cks = make([]*redis.Client, e.NumClients)
	for i := 0; i < e.NumClients; i++ {
		cks[i] = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}

	var done *int32 = new(int32)

	numOps := new(uint64)
	for i := 0; i < e.NumClients; i++ {
		ck1 := cks[i]
		repeat_until_done(func(j int) {
			ck1.Set(ctx, strconv.Itoa(j%e.NumKeys), "somevalue", 0)
			atomic.AddUint64(numOps, 1)
		}, done)
	}

	time.Sleep(e.WarmupTime)
	fmt.Println("Warmup done, starting experiment")
	atomic.StoreUint64(numOps, 0)
	time.Sleep(e.ExperimentTime)
	nOp := atomic.LoadUint64(numOps)
	atomic.StoreInt32(done, 1)
	fmt.Printf("%f puts/sec\n", float64(nOp)/e.ExperimentTime.Seconds())
}

func main() {
	// e := PutThroughputExperiment{NumClients: 10, NumKeys: 100, WarmupTime: 2 * time.Second, ExperimentTime: 10 * time.Second}
	// e.run()
	e := RedisPutThroughputExperiment{NumClients: 10, NumKeys: 100, WarmupTime: 2 * time.Second, ExperimentTime: 10 * time.Second}
	e.run()
	return
}
