package main

import (
	"fmt"
	"time"
)

var DEBUG = false

func Debug(format string, a ...interface{}) {
	if DEBUG {
		debugString := fmt.Sprintf(format, a...)
		fmt.Println(debugString)
	}
}

// Define task types
type Task func()

type PromiseState uint64

const (
	pending   PromiseState = 0 // 0
	fulfilled PromiseState = 1 // 1
	rejected  PromiseState = 2 // 2
)

type PromiseResult[S any, F any] struct {
	successValue  S
	failureValue  F
	nestedPromise *Promise[S, F]
}

// Define a Promise struct
type Promise[S any, F any] struct {
	result    *PromiseResult[S, F]
	err       error
	callbacks []*PromiseCallback[S, F] // Store multiple then callbacks
	state     PromiseState

	// internal for debugging
	id uint64
}

type PromiseCallback[S any, F any] struct {
	onFulfilled func(result *PromiseResult[S, F]) interface{}
	onRejected  func(err error) interface{}
}

// Event loop struct with task queues and a channel for timers
type EventLoop struct {
	taskQueue              []Task
	promisesMicrotaskQueue []Task
	nextTickMicrotaskQueue []Task
	immediateQueue         []Task
	timerChan              chan Task
	activeTimers           uint64
	pendingPromises        uint64
	totalPromises          uint64
	running                bool
}

// Add a task (e.g., regular async tasks)
func AddTask(task Task, eventLoop *EventLoop) {
	eventLoop.taskQueue = append(eventLoop.taskQueue, task)
}

// Add a microtask (e.g., process.nextTick, promises)
func AddPromiseMicrotask(microtask Task, eventLoop *EventLoop) {
	eventLoop.promisesMicrotaskQueue = append(eventLoop.promisesMicrotaskQueue, microtask)
}

func AddNextTickMicroTask(microtask Task, eventLoop *EventLoop) {
	eventLoop.nextTickMicrotaskQueue = append(eventLoop.nextTickMicrotaskQueue, microtask)
}

// Add a task to be run in the immediate phase (setImmediate)
func AddImmediateTask(immediate Task, eventLoop *EventLoop) {
	eventLoop.immediateQueue = append(eventLoop.immediateQueue, immediate)
}

// Non-blocking AddTimerTask using goroutines and a timer channel
func AddTimerTask(task Task, duration time.Duration, eventLoop *EventLoop) {
	Debug("Adding timer task")
	eventLoop.activeTimers++ // Increment active timers
	go func() {
		time.Sleep(duration)
		Debug("Woke up in subroutine")
		eventLoop.timerChan <- task
		Debug("Send to channel in subroutine")
	}()
}

// Process microtasks first (includes promises)
func processMicrotasks(eventLoop *EventLoop) {
	// TODO See if this should jump between queues before going back to main task queue
	// first process nexTick queue entirely
	for len(eventLoop.nextTickMicrotaskQueue) > 0 {
		microtask := eventLoop.nextTickMicrotaskQueue[0]
		eventLoop.nextTickMicrotaskQueue = eventLoop.nextTickMicrotaskQueue[1:]
		microtask()
	}

	// process promises queue entirely
	for len(eventLoop.promisesMicrotaskQueue) > 0 {
		microtask := eventLoop.promisesMicrotaskQueue[0]
		eventLoop.promisesMicrotaskQueue = eventLoop.promisesMicrotaskQueue[1:]
		microtask()
	}
}

func processTimers(eventLoop *EventLoop) {
	Debug("here5")
	shouldContinue := true
	for shouldContinue {
		select {
		case timerTask := <-eventLoop.timerChan:
			// Process timer tasks
			Debug("got a timer task done")
			timerTask()
			processMicrotasks(eventLoop)
			eventLoop.activeTimers-- // Decrement the active timer counter here
			Debug("active timers is now %v", eventLoop.activeTimers)
		default:
			shouldContinue = false
		}
	}
}

// Process immediate tasks (setImmediate phase)
func processImmediateTasks(eventLoop *EventLoop) {
	for len(eventLoop.immediateQueue) > 0 {
		immediateTask := eventLoop.immediateQueue[0]
		eventLoop.immediateQueue = eventLoop.immediateQueue[1:]
		immediateTask()
		processMicrotasks(eventLoop)
	}
}

// Process the main task queue
func processTasks(eventLoop *EventLoop) {
	for len(eventLoop.taskQueue) > 0 {
		task := eventLoop.taskQueue[0]
		eventLoop.taskQueue = eventLoop.taskQueue[1:]
		task()
		processMicrotasks(eventLoop)
	}
}

// Method to resolve the Promise with a result
func resolve[S any, F any](p *Promise[S, F], result *PromiseResult[S, F], eventLoop *EventLoop) {
	// Once the promise is resolved, schedule all .then callbacks in the microtask queue
	Debug("resolving function, num callbacks are %v\n", len(p.callbacks))
	p.result = result
	p.state = fulfilled
	for _, thenFunc := range p.callbacks {
		Debug("Scheduling a microtask")
		AddPromiseMicrotask(func() {
			Debug("removing promise with result %v\n", result)
			Debug("calling a callback")
			thenFunc.onFulfilled(p.result)
		}, eventLoop)
	}

	Debug("resolving promise with id %v", p.id)
	eventLoop.pendingPromises--
}

// Method to resolve the Promise with a result
func reject[S any, F any](p *Promise[S, F], err error, eventLoop *EventLoop) {
	// Once the promise is resolved, schedule all .then callbacks in the microtask queue
	p.err = err
	p.state = rejected
	for _, thenFunc := range p.callbacks {
		AddPromiseMicrotask(func() {
			Debug("removing promise with result %v\n", p.result)
			thenFunc.onRejected(err)
		}, eventLoop)
	}

	Debug("rejecting promise with id %v", p.id)
	eventLoop.pendingPromises--
}

func resolveOrReject[S2 any, F2 any](result *PromiseResult[S2, F2], err error, newPromise *Promise[S2, F2], eventLoop *EventLoop) {
	if err != nil {
		reject(newPromise, err, eventLoop)
	} else if result == nil {
		resolve(newPromise, result, eventLoop)
	} else if result.nestedPromise != nil {
		// TODO Check with node source code here for behavior/think about what makes most sense to put here
		// Could instead check if nestedPromise is already resolved/rejected, and immediately resolve or reject this promise
		then(result.nestedPromise,
			func(result *PromiseResult[S2, F2]) (*PromiseResult[uint64, uint64], error) {
				resolveOrReject(result, nil, newPromise, eventLoop)
				return nil, nil
			}, func(err error) (*PromiseResult[uint64, uint64], error) {
				resolveOrReject(nil, err, newPromise, eventLoop)
				return nil, nil
			}, eventLoop)
	} else {
		resolve(newPromise, result, eventLoop)
	}
}

// Create a new Promise
func NewPromise[S any, F any](executor func(resolveFunc func(result *PromiseResult[S, F]), rejectFunc func(error)), eventLoop *EventLoop) *Promise[S, F] {
	p := &Promise[S, F]{
		id: eventLoop.totalPromises,
	}

	executor(func(result *PromiseResult[S, F]) { resolve(p, result, eventLoop) }, func(err error) { reject(p, err, eventLoop) })
	Debug("Adding promise in constructor with id %v", eventLoop.totalPromises)
	eventLoop.pendingPromises++
	eventLoop.totalPromises++
	return p
}

// Then method to chain multiple callbacks once the promise resolves
func then[S any, F any, S2 any, F2 any](p *Promise[S, F], onFulfilled func(result *PromiseResult[S, F]) (*PromiseResult[S2, F2], error), onRejected func(err error) (*PromiseResult[S2, F2], error), eventLoop *EventLoop) *Promise[S2, F2] {
	// if onRejected == nil {
	// 	onRejected = func(result R) R2 {
	// 		// switch
	// 		return
	// 	}
	// }

	// Create a new Promise for chaining
	newPromise := &Promise[S2, F2]{
		id: eventLoop.totalPromises,
	}

	if p.state == fulfilled {
		AddPromiseMicrotask(func() {
			newResult, err := onFulfilled(p.result)
			resolveOrReject(newResult, err, newPromise, eventLoop)
		}, eventLoop)
	} else if p.state == rejected {
		AddPromiseMicrotask(func() {
			newResult, err := onRejected(p.err)
			resolveOrReject(newResult, err, newPromise, eventLoop)
		}, eventLoop)
	} else {
		p.callbacks = append(p.callbacks, &PromiseCallback[S, F]{
			onFulfilled: func(result *PromiseResult[S, F]) interface{} {
				Debug("resolving promise in then with result %v", result)
				newResult, err := onFulfilled(result)
				Debug("newResult is %v", newResult)
				resolveOrReject(newResult, err, newPromise, eventLoop)
				return nil
			},
			onRejected: func(err error) interface{} {
				newResult, err := onRejected(err)
				resolveOrReject(newResult, err, newPromise, eventLoop)
				return nil
			}})
	}

	Debug("Adding promise in then with id %v", eventLoop.totalPromises)
	eventLoop.pendingPromises++
	eventLoop.totalPromises++
	return newPromise
}

// Start the event loop, incorporating non-blocking timers
func Run(eventLoop *EventLoop) {
	eventLoop.running = true
	for eventLoop.running {
		Debug("Here %v  %v %v %v %v", len(eventLoop.taskQueue), len(eventLoop.immediateQueue), len(eventLoop.promisesMicrotaskQueue), len(eventLoop.nextTickMicrotaskQueue), eventLoop.activeTimers)
		// timer phase
		processTimers(eventLoop)
		// pending callbacks phase
		processTasks(eventLoop)

		// check phase
		processImmediateTasks(eventLoop)

		// If no more tasks and no active timers, stop the loop
		if len(eventLoop.taskQueue) == 0 && len(eventLoop.promisesMicrotaskQueue) == 0 && len(eventLoop.nextTickMicrotaskQueue) == 0 && len(eventLoop.immediateQueue) == 0 && eventLoop.activeTimers == 0 {
			if eventLoop.pendingPromises != 0 {
				fmt.Printf("Execution finished but %v promises still pending\n", eventLoop.pendingPromises)
			}
			Debug("Setting running to false")
			eventLoop.running = false
		}
	}
}

// func setTimeout()

func mainEventLoopFunc(eventLoop *EventLoop) interface{} {
	// Add a task to the event loop
	AddTask(func() {
		fmt.Println("Task 1 regular task")
	}, eventLoop)

	// Create a new Promise that resolves after 200ms
	promise := NewPromise(func(resolveFunc func(result *PromiseResult[string, string]), rejectFunc func(err error)) {
		AddTimerTask(func() {
			fmt.Println("Timer elapsed.")
			resolveFunc(&PromiseResult[string, string]{successValue: "Promise resolved! in time"})
		}, 200*time.Millisecond, eventLoop)
	}, eventLoop)

	then(promise, func(result *PromiseResult[string, string]) (*PromiseResult[string, string], error) {
		AddTask(func() {
			fmt.Printf("Task 2 regular task in Then handler 0: %v\n", result)
		}, eventLoop)
		return nil, nil
	}, nil, eventLoop)

	then(promise, func(result *PromiseResult[string, string]) (*PromiseResult[string, string], error) {
		fmt.Printf("Then handler 1: %v\n", result)
		return nil, nil
	}, nil, eventLoop)

	then(
		// Multiple .then handlers
		then(promise, func(result *PromiseResult[string, string]) (*PromiseResult[string, string], error) {
			// Add a microtask to the event loop once the promise resolves
			AddNextTickMicroTask(func() {
				fmt.Println("Then handler 2:", result)
			}, eventLoop)

			return &PromiseResult[string, string]{successValue: result.successValue + " Chained promise result"}, nil
		}, nil, eventLoop), func(result *PromiseResult[string, string]) (*PromiseResult[string, string], error) {
			fmt.Printf("Then chained handler 3: %v\n", result)
			return nil, nil
		}, nil, eventLoop)

	AddTask(func() {
		fmt.Println("Task 3 regular task")
	}, eventLoop)

	then(promise, func(result *PromiseResult[string, string]) (*PromiseResult[string, string], error) {
		// Add another microtask for the second then handler
		AddNextTickMicroTask(func() {
			fmt.Println("Then chained microtask handler 4:", result)
		}, eventLoop)

		return nil, nil
	}, nil, eventLoop)

	// Add another task
	AddTask(func() {
		fmt.Println("Task 4 regular task")
	}, eventLoop)

	// Add immediate task - ran before above task
	AddImmediateTask(func() {
		fmt.Println("Task 5 immediate task")
	}, eventLoop)

	return nil
}

// Start the event loop
// AddTask(eventLoop, func() {
// mainEventLoopFunc()
// })

// Run(eventLoop)

func main() {
	eventLoop := &EventLoop{
		timerChan:              make(chan Task), // Buffer for timer tasks
		taskQueue:              make([]Task, 0),
		promisesMicrotaskQueue: make([]Task, 0),
		nextTickMicrotaskQueue: make([]Task, 0),
		immediateQueue:         make([]Task, 0),
	}

	AddTask(func() {
		mainEventLoopFunc(eventLoop)
	}, eventLoop)

	// run the runtime!
	Run(eventLoop)
}
