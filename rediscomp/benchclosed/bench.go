package benchclosed

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func runOneClient(initClient func() func(), donePtr *uint64, resetPtr *uint64) int {
	doOp := initClient()

	numOps := 0
	resetCount := uint64(0) // used for warmup

	for atomic.LoadUint64(donePtr) == 0 {
		newResetCount := atomic.LoadUint64(resetPtr)
		if newResetCount > resetCount {
			resetCount = newResetCount
			numOps = 0
		}

		doOp()

		numOps += 1
	}

	return numOps
}

func RunBench(initClient func() func(), numClients int, warmup, runtime time.Duration) {
	done := new(uint64)
	resetCount := new(uint64)

	mu := new(sync.Mutex)
	totalOps := 0

	readyWg := new(sync.WaitGroup)
	doneWg := new(sync.WaitGroup)
	// startClientConnection(done, 0*time.Millisecond)
	for i := 0; i < numClients; i++ {
		readyWg.Add(1)
		doneWg.Add(1)
		go func() {
			numOps := runOneClient(func() func() {
				ret := initClient()
				readyWg.Done()
				return ret
			},
				done, resetCount)
			mu.Lock()
			totalOps += numOps
			mu.Unlock()
			doneWg.Done()
		}()
	}

	// wait for all the clients to be initialized
	readyWg.Wait()

	// wait for warmup
	time.Sleep(warmup)

	// reset the operation counts on all the client goroutines
	atomic.AddUint64(resetCount, 1)

	// actual benchmark time
	time.Sleep(runtime)

	// tell all client threads to stop doing more ops and report back the number
	// of ops they accomplished
	atomic.StoreUint64(done, 1)
	time.Sleep(1 * time.Second)

	doneWg.Wait() // wait for all threads to add in their contribution
	mu.Lock()
	fmt.Printf("Total ops = %d, ops/sec = %f\n", totalOps, float64(totalOps)/runtime.Seconds())
	mu.Unlock()
}
