package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/aof"
	"github.com/mit-pdos/gokv/grove_ffi"
)

func bench_onesize(fname string, writeSize uint64, numThreads uint64) float64 {
	data := make([]byte, writeSize)

	f := aof.CreateAppendOnlyFile(fname)

	warmup := uint64(10)
	n := uint64(100)

	numStillRunning := new(int64)
	atomic.StoreInt64(numStillRunning, int64(numThreads))
	totalTime := new(uint64)
	doneMu := new(sync.Mutex)
	doneCv := sync.NewCond(doneMu)

	doneMu.Lock()
	for i := uint64(0); i < numThreads; i++ {
		go func() {
			for j := uint64(0); j < warmup; j++ {
				f.WaitAppend(f.Append(data))
			}
			start := primitive.TimeNow()
			for j := uint64(0); j < n; j++ {
				f.WaitAppend(f.Append(data))
			}
			end := primitive.TimeNow()

			// first add our time to totalTime and give up ownership of our frac
			// of totalTime
			atomic.AddUint64(totalTime, end-start)

			// indicate that we're done with our main benchmark
			if atomic.AddInt64(numStillRunning, int64(-1)) == 0 {
				// if no one is still running, signal the main thread
				doneMu.Lock()
				doneCv.Signal()
				doneMu.Unlock()
			}

			// wait for everybody to reach this point before moving on
			for atomic.LoadInt64(numStillRunning) != 0 {
				f.WaitAppend(f.Append(data))
			}
		}()
	}

	doneCv.Wait()
	numWritesPerSec := float64(n*numThreads) / (float64(*totalTime) / 1e9)

	return numWritesPerSec
}

func main() {
	sz := uint64(0)
	fname := "test_aof.data"

	for i := 4; i < 20; i += 1 {
		sz = (1 << i)
		grove_ffi.FileWrite(fname, nil)
		fmt.Printf("%d-byte writes -> %f writes/sec\n", sz, bench_onesize(fname, sz, 10))
	}

	for i := 0; i < 20; i += 1 {
		sz += 32 * 1024
		grove_ffi.FileWrite(fname, nil)
		fmt.Printf("%d-byte writes -> %f writes/sec\n", sz, bench_onesize(fname, sz, 10))
	}
}
