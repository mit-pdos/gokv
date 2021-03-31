package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync/atomic"
	"time"
	. "github.com/mit-pdos/gokv/bench"
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
	events []latencyEvent
	// latencies []int64
}

func (l *latencySamples) Write(w io.Writer) {
	for _, e := range l.events {
		fmt.Fprintf(w, "%s, %d\n", e.eventId, e.time)
	}
}

func (l *latencySamples) AddEvent(eventId string) {
	l.events = append(l.events, latencyEvent{time: GetTimestamp(), eventId: eventId})
}

type ValueGenerator interface {
	genValue() []byte
}

type ConstValueGenerator struct {
	Val string
}

func (g *ConstValueGenerator) genValue() []byte {
	return []byte(g.Val)
}

func (g *ConstValueGenerator) String() string {
	return fmt.Sprintf("ConstGen%+v", *g)
}

type RandFixedSizeValueGenerator struct {
	Size int
}

func (g *RandFixedSizeValueGenerator) genValue() []byte {
	b := make([]byte, g.Size)
	_, _ = rand.Read(b)
	return b
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

type GooseKVPutThroughputExperiment struct {
	Rate           float32
	NumKeys        int
	WarmupTime     time.Duration
	ExperimentTime time.Duration
	ValueGenerator ValueGenerator
}

func (e *GooseKVPutThroughputExperiment) run() {
	fmt.Printf("==Testing open loop gokv put throughput with %+v\n", *e)
	p := MakeGooseKVClerkPool(uint64(e.Rate), 100)
	numOps := new(uint64)
	done := new(int32)
	delay := time.Nanosecond * time.Duration(1e9/e.Rate)
	// delay = 100 * time.Nanosecond
	fmt.Printf("%v\n", delay)

	j := 0
	go func() {
		for ; atomic.LoadInt32(done) == 0; j++ {
			// start := time.Now()
			go func(j int) {
				p.Put(uint64(j%e.NumKeys), e.ValueGenerator.genValue())
				atomic.AddUint64(numOps, 1)
			}(j)
			time.Sleep(0)
			// fmt.Println(d)
		}
	}()

	time.Sleep(e.WarmupTime)
	DPrintf("Warmup done, starting experiment")
	atomic.StoreUint64(numOps, 0)
	time.Sleep(e.ExperimentTime)

	numOpsCompleted := atomic.LoadUint64(numOps)
	atomic.StoreInt32(done, 1)
	fmt.Printf("%f puts/sec; %d started\n", float64(numOpsCompleted)/e.ExperimentTime.Seconds(), j)
}

type RedisPutThroughputExperiment struct {
	Rate           float32
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

	cl := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	numOps := new(uint64)
	done := new(int32)
	delay := time.Nanosecond * time.Duration(1e9/e.Rate)
	// delay = 100 * time.Nanosecond
	fmt.Printf("%v\n", delay)
	j := 0
	go func() {
		for ; atomic.LoadInt32(done) == 0; j++ {
			go func(j int) {
				cl.Set(ctx, strconv.Itoa(j%e.NumKeys), e.ValueGenerator.genValue(), 0)
				atomic.AddUint64(numOps, 1)
			}(j)
			time.Sleep(delay)
		}
	}()

	time.Sleep(e.WarmupTime)
	DPrintf("Warmup done, starting experiment")
	atomic.StoreUint64(numOps, 0)
	time.Sleep(e.ExperimentTime)

	numOpsCompleted := atomic.LoadUint64(numOps)
	atomic.StoreInt32(done, 1)
	fmt.Printf("%f puts/sec; %d started\n", float64(numOpsCompleted)/e.ExperimentTime.Seconds(), j)
}

type Experiment interface {
	run()
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	runtime.GOMAXPROCS(2)

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	for _, e := range experiments {
		e.run()
	}
	return
}
