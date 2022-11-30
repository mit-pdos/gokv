package aof

import (
	// "log"
	"sync"
	// "time"

	"github.com/mit-pdos/gokv/grove_ffi"
	// "github.com/tchajed/goose/machine"
)

type AppendOnlyFile struct {
	fname string
	mu    *sync.Mutex

	durableCond *sync.Cond
	lengthCond  *sync.Cond

	membuf        []byte
	length        uint64 // logical length
	durableLength uint64
}

func CreateAppendOnlyFile(fname string) *AppendOnlyFile {
	a := new(AppendOnlyFile)
	a.mu = new(sync.Mutex)
	a.lengthCond = sync.NewCond(a.mu)
	a.durableCond = sync.NewCond(a.mu)

	go func() {
		a.mu.Lock()
		// lastPrinted := time.Now()
		for {
			// TODO: also check how often we end up having to wait here right
			// after completing AtomicAppend()
			if len(a.membuf) == 0 {
				// begin := time.Now()
				a.lengthCond.Wait()
				// end := time.Now()
				// if end.After(lastPrinted.Add(100 * time.Millisecond)) {
				// diff := end.Sub(begin)
				// log.Printf("aof.BackgroundThread: waited for: %v for %d bytes\n", diff, len(a.membuf))
				// lastPrinted = end
				// }
				continue
			}

			l := a.membuf
			newLength := a.length
			a.membuf = make([]byte, 0)
			a.mu.Unlock()
			// log.Printf("AtomicAppend %d bytes\n", len(l))

			grove_ffi.FileAppend(fname, l)

			a.mu.Lock()
			a.durableLength = newLength
			a.durableCond.Broadcast()
			continue
		}
	}()

	return a
}

// var beginTime = time.Now()
// var appendStartTimes []time.Duration

func (a *AppendOnlyFile) Append(data []byte) uint64 {
	a.mu.Lock()
	// timeSinceLastCall := time.Now().Sub(beginTime)
	// appendStartTimes = append(appendStartTimes, timeSinceLastCall)
	// beginTime = time.Now()
	// if len(appendStartTimes) >= 128 {
	// if machine.RandomUint64()%64 == 0 {
	// log.Printf("%v\n", appendStartTimes)
	// }
	// appendStartTimes = nil
	//}

	// log.Printf("Append %d bytes\n", len(data))

	a.membuf = append(a.membuf, data...)
	for a.length+uint64(len(data)) < a.length {
	}

	a.length = a.length + uint64(len(data))
	r := a.length
	a.lengthCond.Signal()
	a.mu.Unlock()
	// if machine.RandomUint64()%1024 == 0 {
	// critSectionTime := time.Now().Sub(beginTime)
	// log.Printf("aof.Append() time since last: %v\n", timeSinceLastCall)
	// log.Printf("aof.Append() crit section: %v\n", critSectionTime)
	// }
	return r
}

func (a *AppendOnlyFile) Close() {
	panic("unimpl")
}

func (a *AppendOnlyFile) WaitAppend(length uint64) {
	a.mu.Lock()
	for a.durableLength < length {
		a.durableCond.Wait()
	}
	a.mu.Unlock()
}
