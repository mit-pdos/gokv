package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/upamanyu/gokv"
	"io"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
	"os"
)

// XXX: this doesn't use monotonic time.
func GetTimestamp() int64 {
	return time.Now().UnixNano()
}

type latencyEvent struct {
	time    int64
	eventId string
}

type latencySamples struct {
	events  []latencyEvent
	// latencies []int64
}

func (l *latencySamples) Write(w io.Writer) {
	for _, e := range l.events {
		fmt.Fprintf(w, "%s, %d\n", e.eventId, e.time)
	}
}

func (l *latencySamples) AddEvent(eventId string) {
	l.events = append(l.events, latencyEvent{time:GetTimestamp(), eventId:eventId})
}

type ValueGenerator interface {
	genValue() string
}

type ConstValueGenerator struct {
	Val string
}

func (g *ConstValueGenerator) genValue() string {
	return g.Val
}

func (g *ConstValueGenerator) String() string {
	return fmt.Sprintf("ConstGen%+v", *g)
}

type RandFixedSizeValueGenerator struct {
	Size int
}

func (g *RandFixedSizeValueGenerator) genValue() string {
	b := make([]byte, g.Size)
	_, _ = rand.Read(b)
	return string(b)
}

func (g *RandFixedSizeValueGenerator) String() string {
	return fmt.Sprintf("FixedSizeGen%+v", *g)
}

func repeat_until_done(f func(int), done *int32) {
	go func() {
		for i := 0; atomic.LoadInt32(done) == 0; i++ {
			f(i)
		}
	}()
}

// TODO: make the output files a parameter
type PutThroughputExperiment struct {
	NumClients     int
	NumKeys        int
	WarmupTime     time.Duration
	ExperimentTime time.Duration
	ValueGenerator ValueGenerator
}

func (e *PutThroughputExperiment) run() {
	fmt.Printf("==Testing gokv put throughput with %+v\n", *e)

	// make clerks
	var cks []*gokv.GoKVClerk
	var lss []*latencySamples
	cks = make([]*gokv.GoKVClerk, e.NumClients)
	lss = make([]*latencySamples, e.NumClients)
	for i := 0; i < e.NumClients; i++ {
		cks[i] = gokv.MakeKVClerk(uint64(i), "localhost")
		lss[i] = &latencySamples{nil}
	}

	var done *int32 = new(int32)

	numOps := new(uint64)
	for i := 0; i < e.NumClients; i++ {
		ck := cks[i]
		ls := lss[i]
		repeat_until_done(func(j int) {
			ls.AddEvent("PutBeg")
			ck.Put(uint64(j%e.NumKeys), e.ValueGenerator.genValue())
			ls.AddEvent("PutEnd")
			atomic.AddUint64(numOps, 1)
		}, done)
	}

	time.Sleep(e.WarmupTime)

	DPrintf("Warmup done, starting experiment")
	for i := range lss {
		*lss[i] = latencySamples{nil}
	}

	atomic.StoreUint64(numOps, 0)
	time.Sleep(e.ExperimentTime)
	nOp := atomic.LoadUint64(numOps)
	atomic.StoreInt32(done, 1)
	fmt.Printf("%f puts/sec\n", float64(nOp)/e.ExperimentTime.Seconds())

	f, err := os.Create(fmt.Sprintf("data/put_thruput_%d.txt", GetTimestamp()/1e6))
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(f, "PutThruput%+v\n", *e)
	for _, ls := range lss {
		ls.Write(f)
	}
	f.Close()
}

type RedisPutThroughputExperiment struct {
	NumClients     int
	NumKeys        int
	WarmupTime     time.Duration
	ExperimentTime time.Duration
	ValueGenerator ValueGenerator
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
	var lss []*latencySamples
	cks = make([]*redis.Client, e.NumClients)
	lss = make([]*latencySamples, e.NumClients)
	for i := 0; i < e.NumClients; i++ {
		cks[i] = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
		lss[i] = &latencySamples{nil}
	}

	var done *int32 = new(int32)

	numOps := new(uint64)
	for i := 0; i < e.NumClients; i++ {
		ck1 := cks[i]
		ls := lss[i]
		repeat_until_done(func(j int) {
			ls.AddEvent("RPutBeg")
			ck1.Set(ctx, strconv.Itoa(j%e.NumKeys), e.ValueGenerator.genValue(), 0)
			ls.AddEvent("RPutEnd")
			atomic.AddUint64(numOps, 1)
		}, done)
	}

	time.Sleep(e.WarmupTime)
	DPrintf("Warmup done, starting experiment")
	for i := range lss {
		*lss[i] = latencySamples{nil}
	}
	atomic.StoreUint64(numOps, 0)

	time.Sleep(e.ExperimentTime)
	nOp := atomic.LoadUint64(numOps)
	atomic.StoreInt32(done, 1)
	fmt.Printf("%f puts/sec\n", float64(nOp)/e.ExperimentTime.Seconds())

	f, err := os.Create(fmt.Sprintf("data/redis_put_thruput_%d.txt", GetTimestamp()/1e6))
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(f, "RedisPutThruput%+v\n", *e)
	for _, ls := range lss {
		ls.Write(f)
	}
	f.Close()
}

type Experiment interface {
	run()
}

func main() {
	for _, e := range experiments {
		e.run()
	}
	return
}
