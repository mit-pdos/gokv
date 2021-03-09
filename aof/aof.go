package aof

import (
	"sync"
	"github.com/mit-pdos/lockservice/grove_ffi"
)

type AppendOnlyFile struct {
	fname string
	mu *sync.Mutex

	durableCond *sync.Cond
	lengthCond *sync.Cond

	membuf []byte
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
			a.membuf = make([]byte, 0)
			a.mu.Unlock()

			grove_ffi.AtomicAppend(fname, l)

			a.mu.Lock()
			a.durableLength += uint64(len(l))
			a.durableCond.Broadcast()
		}
	}()

	return a
}

func (a *AppendOnlyFile) Append(data []byte) uint64 {
	a.mu.Lock()
	a.membuf = append(a.membuf, data...)
	r := a.durableLength + uint64(len(a.membuf))
	a.lengthCond.Signal()
	a.mu.Unlock()
	return r
}

func (a *AppendOnlyFile) WaitAppend(length uint64) {
	a.mu.Lock()
	for a.durableLength < length {
		a.durableCond.Wait()
	}
	a.mu.Unlock()
}
