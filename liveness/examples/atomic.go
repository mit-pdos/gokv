package examples

import (
	"sync"
)

type AtomicUint64 struct {
	mu *sync.Mutex
	v  uint64
}

func MakeAtomicUint64() *AtomicUint64 {
	return &AtomicUint64{mu: new(sync.Mutex), v: 0}
}

func (a *AtomicUint64) Write(v uint64) {
	a.mu.Lock()
	a.v = v
	a.mu.Unlock()
}

func (a *AtomicUint64) Read() uint64 {
	a.mu.Lock()
	v := a.v
	a.mu.Unlock()
	return v
}
