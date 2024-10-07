package main

import (
	"testing"
	"time"

	"github.com/goose-lang/primitive"
)

func TestTimersBeforeImmediate(t *testing.T) {
	eventLoop := &EventLoop{
		timerChan:              make(chan Task), // Buffer for timer tasks
		taskQueue:              make([]Task, 0),
		promisesMicrotaskQueue: make([]Task, 0),
		nextTickMicrotaskQueue: make([]Task, 0),
		immediateQueue:         make([]Task, 0),
	}

	returnValues := make([]int, 0)

	setImmediate(func() {
		returnValues = append(returnValues, 2)
	}, eventLoop)

	setTimeout(func() {
		returnValues = append(returnValues, 1)
	}, 0*time.Millisecond, eventLoop)

	// Sleep to make sure timer task is on the queue
	time.Sleep(10 * time.Millisecond)

	// run the runtime!
	Run(eventLoop)

	primitive.Assert(returnValues[0] == 1 && returnValues[1] == 2)
}

func TestNextTickBeforePromise(t *testing.T) {
	eventLoop := &EventLoop{
		timerChan:              make(chan Task), // Buffer for timer tasks
		taskQueue:              make([]Task, 0),
		promisesMicrotaskQueue: make([]Task, 0),
		nextTickMicrotaskQueue: make([]Task, 0),
		immediateQueue:         make([]Task, 0),
	}

	returnValues := make([]int, 0)

	// queue a macrotask to enable microtask processing
	setImmediate(func() {}, eventLoop)

	queueMicrotask(func() {
		returnValues = append(returnValues, 2)
	}, eventLoop)

	then(PromiseResolved(uint64(1), eventLoop), func(result *PromiseResult[uint64, uint64]) (*PromiseResult[uint64, uint64], error) {
		returnValues = append(returnValues, 3)
		return nil, nil
	}, nil, eventLoop)

	processNextTick(func() {
		returnValues = append(returnValues, 1)
	}, eventLoop)

	// run the runtime!
	Run(eventLoop)

	primitive.Assert(returnValues[0] == 1 && returnValues[1] == 2 && returnValues[2] == 3)
}

func TestJumpBetweenPromisesAndNextTickBeforeNextTick(t *testing.T) {
	eventLoop := &EventLoop{
		timerChan:              make(chan Task), // Buffer for timer tasks
		taskQueue:              make([]Task, 0),
		promisesMicrotaskQueue: make([]Task, 0),
		nextTickMicrotaskQueue: make([]Task, 0),
		immediateQueue:         make([]Task, 0),
	}

	returnValues := make([]int, 0)

	// queue a macrotask to enable microtask processing
	setImmediate(func() {
		returnValues = append(returnValues, 1)
	}, eventLoop)

	queueMicrotask(func() {
		returnValues = append(returnValues, 2)
		processNextTick(func() {
			returnValues = append(returnValues, 3)
			then(PromiseResolved(uint64(1), eventLoop), func(result *PromiseResult[uint64, uint64]) (*PromiseResult[uint64, uint64], error) {
				returnValues = append(returnValues, 4)
				return nil, nil
			}, nil, eventLoop)
		}, eventLoop)
	}, eventLoop)

	// TODO it would be nice to know that using 0 synchronously puts the callback on the queue, see
	// if node does this already.
	setImmediate(func() {
		returnValues = append(returnValues, 5)
	}, eventLoop)

	// run the runtime!
	Run(eventLoop)

	primitive.Assert(returnValues[0] == 1 && returnValues[1] == 2 && returnValues[2] == 3 && returnValues[3] == 4 && returnValues[4] == 5)
}

func TestProcessAllMicrotasksBeforeSwitchingQueues(t *testing.T) {
	eventLoop := &EventLoop{
		timerChan:              make(chan Task), // Buffer for timer tasks
		taskQueue:              make([]Task, 0),
		promisesMicrotaskQueue: make([]Task, 0),
		nextTickMicrotaskQueue: make([]Task, 0),
		immediateQueue:         make([]Task, 0),
	}

	returnValues := make([]int, 0)

	// queue a macrotask to enable microtask processing
	setImmediate(func() {}, eventLoop)

	queueMicrotask(func() {
		returnValues = append(returnValues, 1)
		processNextTick(func() {
			returnValues = append(returnValues, 3)
		}, eventLoop)
	}, eventLoop)

	then(PromiseResolved(uint64(1), eventLoop), func(result *PromiseResult[uint64, uint64]) (*PromiseResult[uint64, uint64], error) {
		returnValues = append(returnValues, 2)
		return nil, nil
	}, nil, eventLoop)

	setImmediate(func() {
		returnValues = append(returnValues, 4)
	}, eventLoop)

	// run the runtime!
	Run(eventLoop)

	primitive.Assert(returnValues[0] == 1 && returnValues[1] == 2 && returnValues[2] == 3 && returnValues[3] == 4)
}
