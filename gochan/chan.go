package gochan

import (
	"sync"
)

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

type chanstate[T any] struct {
	cap      int
	closed   bool
	sent     []T
	received int
}

type Channel[T any] struct {
	mu   sync.Mutex
	cond *sync.Cond
	chanstate[T]
}

func (ch *Channel[T]) Cap() int {
	if ch == nil {
		return 0
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return ch.cap
}

func (ch *Channel[T]) Len() int {
	if ch == nil {
		return 0
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	return min(max(len(ch.sent)-ch.received, 0), ch.cap)
}

func NewChan[T any](cap int) *Channel[T] {
	ch := &Channel[T]{}
	ch.cond = sync.NewCond(&ch.mu)
	ch.cap = cap
	return ch
}

func (ch *Channel[T]) Send(v T) {
	for ch == nil {
	}
	ch.mu.Lock()
	assert(!ch.closed, "send on closed channel")
	i := int(len(ch.sent))
	ch.sent = append(ch.sent, v)
	ch.cond.Broadcast()

	for !(i < ch.received+ch.cap) {
		ch.cond.Wait()
	}
	ch.mu.Unlock()
}

func (ch *Channel[T]) Receive() (T, bool) {
	for ch == nil {
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	i := ch.received
	ch.received++
	ch.cond.Broadcast()

	for {
		if i < len(ch.sent) {
			break
		} else if ch.closed {
			var zero T
			return zero, false
		}
		ch.cond.Wait()
	}
	v := ch.sent[i]
	return v, true
}

func (ch *Channel[T]) Close() {
	assert(ch != nil, "close of nil channel")
	ch.mu.Lock()
	defer ch.mu.Unlock()
	assert(!ch.closed, "close of closed channel")
	ch.closed = true
	ch.cond.Broadcast()
}

func (ch *Channel[T]) NonblockingSend(v T) bool {
	if ch == nil {
		return false
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	assert(!ch.closed, "send on closed channel")
	if len(ch.sent) < ch.received+ch.cap {
		ch.sent = append(ch.sent, v)
		ch.cond.Broadcast()
		return true
	} else {
		return false
	}
}

func (ch *Channel[T]) NonblockingReceive() (bool, T, bool) {
	var zero T
	if ch == nil {
		return false, zero, false
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if ch.received < len(ch.sent) {
		ch.received++
		ch.cond.Broadcast()
		return true, ch.sent[ch.received-1], true
	} else if ch.closed {
		return true, zero, false
	} else {
		return false, zero, false
	}
}
