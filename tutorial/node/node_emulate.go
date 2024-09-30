package main

import (
	"fmt"
	"time"
)

// Define task types
type Task func()

type PromiseState uint64

const (
	pending   PromiseState = 0 // 0
	fulfilled PromiseState = 1 // 1
	rejected  PromiseState = 2 // 2
)

var eventLoop = &EventLoop{
	timerChan:      make(chan Task, 10), // Buffer for timer tasks
	taskQueue:      make([]Task, 0),
	microtaskQueue: make([]Task, 0),
	immediateQueue: make([]Task, 0),
}

// Define a Promise struct
type Promise struct {
	result    interface{}
	callbacks []*PromiseCallback // Store multiple then callbacks
	state     PromiseState
}

type PromiseCallback struct {
	onFulfilled func(result interface{}) interface{}
	onRejected  func(result interface{}) interface{}
}

// Create a new Promise
func NewPromise(executor func(resolveFunc func(interface{}), rejectFunc func(interface{}))) *Promise {
	p := &Promise{}
	executor(p.resolve, p.reject)
	// fmt.Println("Adding promise in constructor")
	eventLoop.pendingPromises++
	return p
}

// Method to resolve the Promise with a result
func (p *Promise) resolve(result interface{}) {
	// Once the promise is resolved, schedule all .then callbacks in the microtask queue
	// fmt.Printf("resolving function, num callbacks are %v\n", len(p.callbacks))
	p.result = result
	p.state = fulfilled
	for _, thenFunc := range p.callbacks {
		// fmt.Println("Scheduling a microtask")
		eventLoop.AddMicrotask(func() {
			// fmt.Printf("removing promise with result %v\n", result)
			// fmt.Println("calling a callback")
			thenFunc.onFulfilled(result)
		})
	}

	eventLoop.pendingPromises--
}

// Method to resolve the Promise with a result
func (p *Promise) reject(result interface{}) {
	// Once the promise is resolved, schedule all .then callbacks in the microtask queue
	p.result = result
	p.state = rejected
	for _, thenFunc := range p.callbacks {
		eventLoop.AddMicrotask(func() {
			// fmt.Printf("removing promise with result %v\n", result)
			thenFunc.onRejected(result)
		})
	}

	eventLoop.pendingPromises--
}

// Then method to chain multiple callbacks once the promise resolves
func (p *Promise) then(onFulfilled func(result interface{}) interface{}, onRejected func(result interface{}) interface{}) *Promise {
	if onRejected == nil {
		onRejected = func(result interface{}) interface{} {
			return result
		}
	}

	// Create a new Promise for chaining
	newPromise := &Promise{}

	if p.state == fulfilled {
		eventLoop.AddMicrotask(func() {
			newPromise.resolve(onFulfilled(p.result))
		})
	} else if p.state == rejected {
		eventLoop.AddMicrotask(func() {
			newPromise.reject(onRejected(p.result))
		})
	} else {
		p.callbacks = append(p.callbacks, &PromiseCallback{onFulfilled: func(result interface{}) interface{} {
			// fmt.Println("resolving promise in then")
			newPromise.resolve(onFulfilled(result))
			return nil
		}, onRejected: func(result interface{}) interface{} {
			newPromise.reject(onRejected(result))
			return nil
		}})
	}

	// fmt.Println("Adding promise in then")
	eventLoop.pendingPromises++
	return newPromise
}

// Event loop struct with task queues and a channel for timers
type EventLoop struct {
	taskQueue       []Task
	microtaskQueue  []Task
	immediateQueue  []Task
	timerChan       chan Task
	activeTimers    uint64
	pendingPromises uint64
	running         bool
}

// Add a task (e.g., regular async tasks)
func (el *EventLoop) AddTask(task Task) {
	el.taskQueue = append(el.taskQueue, task)
}

// Add a microtask (e.g., process.nextTick, promises)
func (el *EventLoop) AddMicrotask(microtask Task) {
	el.microtaskQueue = append(el.microtaskQueue, microtask)
}

// Add a task to be run in the immediate phase (setImmediate)
func (el *EventLoop) AddImmediateTask(immediate Task) {
	el.immediateQueue = append(el.immediateQueue, immediate)
}

// Non-blocking AddTimerTask using goroutines and a timer channel
func (el *EventLoop) AddTimerTask(task Task, duration time.Duration) {
	// fmt.Println("Adding timer task")
	el.activeTimers++ // Increment active timers
	go func() {
		// fmt.Println("GOing to sleep in subroutine")
		time.Sleep(duration)
		// fmt.Println("Woke up in subroutine")
		el.timerChan <- task
		// fmt.Println("Send to channel in subroutine")
	}()
}

// Process microtasks first (includes promises)
func (el *EventLoop) processMicrotasks() {
	for len(el.microtaskQueue) > 0 {
		microtask := el.microtaskQueue[0]
		el.microtaskQueue = el.microtaskQueue[1:]
		microtask()
	}
}

// Process immediate tasks (setImmediate phase)
func (el *EventLoop) processImmediateTask() {
	if len(el.immediateQueue) == 0 {
		return
	}
	immediateTask := el.immediateQueue[0]
	el.immediateQueue = el.immediateQueue[1:]
	immediateTask()
}

// Process the main task queue
func (el *EventLoop) processTask() {
	if len(el.taskQueue) == 0 {
		return
	}

	task := el.taskQueue[0]
	el.taskQueue = el.taskQueue[1:]
	task()
}

// Start the event loop, incorporating non-blocking timers
func (el *EventLoop) Run() {
	el.running = true
	for el.running {
		// Non-blocking select for timer and task processing
		select {
		case timerTask := <-el.timerChan:
			// Process timer tasks
			// fmt.Println("got a timer task done")
			timerTask()
			el.activeTimers-- // Decrement the active timer counter here
			// fmt.Printf("active timers is now %v", el.activeTimers)

		default:
			// Process regular tasks and immediate tasks
			el.processMicrotasks()
			el.processTask()
			el.processImmediateTask()
		}

		// If no more tasks and no active timers, stop the loop
		if len(el.taskQueue) == 0 && len(el.microtaskQueue) == 0 && len(el.immediateQueue) == 0 && el.activeTimers == 0 {
			if eventLoop.pendingPromises != 0 {
				fmt.Printf("Execution finished but %v promises still pending\n", eventLoop.pendingPromises)
			}
			// fmt.Println("Setting running to false")
			el.running = false
		}
	}
}

// func setTimeout()

func mainEventLoopFunc() interface{} {
	// Add a task to the event loop
	eventLoop.AddTask(func() {
		fmt.Println("Task 1")
	})

	// Create a new Promise that resolves after 200ms
	promise := NewPromise(func(resolveFunc func(interface{}), rejectFunc func(interface{})) {
		eventLoop.AddTimerTask(func() {
			resolveFunc("Promise resolved! in timer.")
		}, 200*time.Millisecond)
	})

	promise.then(func(result interface{}) interface{} {
		eventLoop.AddTask(func() {
			fmt.Printf("Task 2 Then handler 0: %v\n", result)
		})
		return nil
	}, nil)

	promise.then(func(result interface{}) interface{} {
		fmt.Printf("Then handler 1: %v\n", result)
		return nil
	}, nil)

	// Multiple .then handlers
	promise.then(func(result interface{}) interface{} {
		// Add a microtask to the event loop once the promise resolves
		eventLoop.AddMicrotask(func() {
			fmt.Println("Then handler 2:", result)
		})

		resultString, ok := result.(string)
		if !ok {
			return " Could not chain promise result, add err handling!"
		}
		return resultString + " Chained promise result"
	}, nil).then(func(result interface{}) interface{} {
		fmt.Printf("Then chained handler 3: %v\n", result)
		return nil
	}, nil)

	eventLoop.AddTask(func() {
		fmt.Println("Task 3")
	})

	promise.then(func(result interface{}) interface{} {
		// Add another microtask for the second then handler
		eventLoop.AddMicrotask(func() {
			fmt.Println("Then handler 4:", result)
		})

		return nil
	}, nil)

	// Add another task
	eventLoop.AddTask(func() {
		fmt.Println("Task 4")
	})

	return nil
}

func main() {

	// Start the event loop
	eventLoop.AddTask(func() {
		mainEventLoopFunc()
	})

	eventLoop.Run()
}
