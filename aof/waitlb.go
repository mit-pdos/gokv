package aof

import "sync"

type waitEntry struct {
	cond *sync.Cond
	lb   uint64
}

// queue of waitEntries
type waitEntryQ struct {
	conds []waitEntry
	start uint64
	end   uint64
}

func enqueueWaiter() {
	panic("impl")
}

func dequeueWaiter() {
	panic("impl")
}

// XXX: right now, this allocates more and more entries.
type WaitLowerbound struct {
	mu      *sync.Mutex
	v       uint64
	waiters []waitEntry
	// activePriority <= waiters[0].priority
}

// precond: priority is bigger than any priority registered previously.
func (wlb *WaitLowerbound) GetWaiter(lb uint64) func() {
	wlb.mu.Lock()
	cond := sync.NewCond(wlb.mu)
	wlb.waiters = append(wlb.waiters, waitEntry{cond: cond, lb: lb})
	wlb.mu.Unlock()
	return func() {
		wlb.mu.Lock()
		for lb < wlb.v {
			cond.Wait()
		}
		wlb.mu.Unlock()
	}
}

// requires v to be greater than whatever was set previously
func (wlb *WaitLowerbound) Set(v uint64) {
	wlb.mu.Lock()
	var i = uint64(0)
	for i < uint64(len(wlb.waiters)) {
		w := wlb.waiters[i]
		if w.lb <= v {
			w.cond.Signal()
		} else {
			break
		}

		i += 1
	}
	// invariant: wlb.waiters[0:i] are below wlb.v, and we can get rid of those
	wlb.waiters = wlb.waiters[i:]

	wlb.mu.Unlock()
}
