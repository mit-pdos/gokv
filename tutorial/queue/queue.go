package queue

import (
	"sync"
)

type Queue struct {
	queue []uint64
	cond  *sync.Cond
	lock  *sync.Mutex
	first uint64
	count uint64
}

func NewQueue(queue_size uint64) Queue {
	lock := new(sync.Mutex)
	return Queue{
		queue: make([]uint64, queue_size),
		cond:  sync.NewCond(lock),
		lock:  lock,
		first: 0,
		count: 0,
	}
}

func NewQueueRef(queue_size uint64) *Queue {
	lock := new(sync.Mutex)
	return &Queue{
		queue: make([]uint64, queue_size),
		cond:  sync.NewCond(lock),
		lock:  lock,
		first: 0,
		count: 0,
	}
}

func (q *Queue) Enqueue(a uint64) {
	q.lock.Lock()
	var queue_size uint64 = uint64(len(q.queue))
	for q.count >= queue_size {
		q.cond.Wait()
	}
	var last uint64 = (q.first + q.count) % queue_size
	q.queue[last] = a
	q.count += 1
	q.lock.Unlock()
	q.cond.Broadcast()
}

func (q *Queue) Dequeue() uint64 {
	q.lock.Lock()
	var queue_size uint64 = uint64(len(q.queue))
	for q.count == 0 {
		q.cond.Wait()
	}
	res := q.queue[q.first]
	q.first = (q.first + 1) % queue_size
	q.count -= 1
	q.lock.Unlock()
	q.cond.Broadcast()
	return res
}

func (q *Queue) Peek() (uint64, bool) {
	q.lock.Lock()
	if q.count > 0 {
		first := q.queue[q.first]
		q.lock.Unlock()
		return first, true
	}
	q.lock.Unlock()
	return 0, false
}
