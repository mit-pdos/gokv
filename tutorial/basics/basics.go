package basics

import (
	"sync"
)

type Tracker struct {
	mu *sync.Mutex
	m  map[uint64]uint64
}

func (t *Tracker) lookup_locked(k uint64) (uint64, bool) {
	v, ok := t.m[k]
	return v, ok
}

func (t *Tracker) register_locked(k uint64, v uint64) bool {
	_, ok := t.lookup_locked(k)
	if ok {
		return false
	}

	t.m[k] = v
	return true
}

func (t *Tracker) Lookup(k uint64) (uint64, bool) {
	t.mu.Lock()
	v, ok := t.lookup_locked(k)
	t.mu.Unlock()
	return v, ok
}

func (t *Tracker) Register(k uint64, v uint64) bool {
	t.mu.Lock()
	r := t.register_locked(k, v)
	t.mu.Unlock()
	return r
}

func MakeTracker() *Tracker {
	t := new(Tracker)
	t.mu = new(sync.Mutex)
	t.m = make(map[uint64]uint64)
	return t
}
