package aof

import (
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

type AppendOnlyFile struct {
	mu *sync.Mutex

	// Having two condvars helps make sure that the FileAppend() thread only
	// wakes up the Wait() threads that are actually done.
	oldDurableCond *sync.Cond
	durableCond    *sync.Cond
	lengthCond     *sync.Cond

	membuf        []byte
	length        uint64 // logical length
	durableLength uint64

	closeRequested bool
	closed         bool
	closedCond     *sync.Cond
}

func CreateAppendOnlyFile(fname string) *AppendOnlyFile {
	a := new(AppendOnlyFile)
	a.mu = new(sync.Mutex)
	a.lengthCond = sync.NewCond(a.mu)
	a.oldDurableCond = sync.NewCond(a.mu)
	a.durableCond = sync.NewCond(a.mu)
	a.closedCond = sync.NewCond(a.mu)

	go func() {
		a.mu.Lock()
		for {
			if len(a.membuf) == 0 && !a.closeRequested {
				a.lengthCond.Wait()
				continue
			}

			if a.closeRequested {
				// Write the remaining stuff so that we can wake up anyone
				// that's already waiting
				grove_ffi.FileAppend(fname, a.membuf)
				a.membuf = make([]byte, 0)
				a.durableLength = a.length
				a.durableCond.Broadcast()

				a.closed = true
				a.closedCond.Broadcast()
				a.mu.Unlock()
				break
			}

			l := a.membuf
			newLength := a.length
			a.membuf = make([]byte, 0)
			cond := a.durableCond

			// swap the condvars, oldDurableCond is effectively unused at this
			// point
			a.durableCond = a.oldDurableCond
			a.oldDurableCond = cond

			a.mu.Unlock()

			grove_ffi.FileAppend(fname, l)

			a.mu.Lock()
			a.durableLength = newLength
			cond.Broadcast()
			continue
		}
	}()

	return a
}

// NOTE: cannot be called concurrently with Append()
func (a *AppendOnlyFile) Close() {
	a.mu.Lock()
	a.closeRequested = true
	a.lengthCond.Signal()
	for !a.closed {
		a.closedCond.Wait()
	}
	a.mu.Unlock()
}

// NOTE: cannot be called concurrently with Close()
func (a *AppendOnlyFile) Append(data []byte) uint64 {
	a.mu.Lock()

	// XXX: using WriteBytes instead of append() because Goose has no reasoning
	// principles for SliceAppend
	a.membuf = marshal.WriteBytes(a.membuf, data)
	a.length = std.SumAssumeNoOverflow(a.length, uint64(len(data)))
	r := a.length
	a.lengthCond.Signal()

	a.mu.Unlock()
	return r
}

func (a *AppendOnlyFile) WaitAppend(length uint64) {
	a.mu.Lock()
	var cond *sync.Cond

	if length + uint64(len(a.membuf)) < a.length {
		cond = a.oldDurableCond // XXX: wait for data the FileAppend thread already started writing
	} else {
		cond = a.durableCond // wait for data that's still in membuf, associated with the new condvar
	}

	for a.durableLength < length {
		cond.Wait()
	}
	a.mu.Unlock()
}
