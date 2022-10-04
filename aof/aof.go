package aof

import (
	// "log"
	"sync"

	"github.com/mit-pdos/lockservice/grove_ffi"
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
		for {
			if len(a.membuf) == 0 {
				a.lengthCond.Wait()
				continue
			}

			l := a.membuf
			newLength := a.length
			a.membuf = make([]byte, 0)
			a.mu.Unlock()
			// log.Printf("AtomicAppend %d bytes\n", len(l))

			grove_ffi.AtomicAppend(fname, l)

			a.mu.Lock()
			a.durableLength = newLength
			a.durableCond.Broadcast()
			continue
		}
	}()

	return a
}

func (a *AppendOnlyFile) Append(data []byte) uint64 {
	a.mu.Lock()
	a.membuf = append(a.membuf, data...)
	for a.length+uint64(len(data)) < a.length {
	}

	a.length = a.length + uint64(len(data))
	r := a.length
	a.lengthCond.Signal()
	a.mu.Unlock()
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
