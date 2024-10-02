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

type Runtime struct {
	eventLoop *EventLoop
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
func (r *Runtime) AddTask(task Task) {
	r.eventLoop.taskQueue = append(r.eventLoop.taskQueue, task)
}

// Add a microtask (e.g., process.nextTick, promises)
func (r *Runtime) AddMicrotask(microtask Task) {
	r.eventLoop.microtaskQueue = append(r.eventLoop.microtaskQueue, microtask)
}

// Add a task to be run in the immediate phase (setImmediate)
func (r *Runtime) AddImmediateTask(immediate Task) {
	r.eventLoop.immediateQueue = append(r.eventLoop.immediateQueue, immediate)
}

// Non-blocking AddTimerTask using goroutines and a timer channel
func (r *Runtime) AddTimerTask(task Task, duration time.Duration) {
	// fmt.Println("Adding timer task")
	r.eventLoop.activeTimers++ // Increment active timers
	go func() {
		time.Sleep(duration)
		// fmt.Println("Woke up in subroutine")
		r.eventLoop.timerChan <- task
		// fmt.Println("Send to channel in subroutine")
	}()
}

// Process microtasks first (includes promises)
func (r *Runtime) processMicrotasks() {
	for len(r.eventLoop.microtaskQueue) > 0 {
		microtask := r.eventLoop.microtaskQueue[0]
		r.eventLoop.microtaskQueue = r.eventLoop.microtaskQueue[1:]
		microtask()
	}
}

func (r *Runtime) processTimers() {
Loop:
	for {
		select {
		case timerTask := <-r.eventLoop.timerChan:
			// Process timer tasks
			// fmt.Println("got a timer task done")
			timerTask()
			r.processMicrotasks()
			r.eventLoop.activeTimers-- // Decrement the active timer counter here
			// fmt.Printf("active timers is now %v", r.eventLoop.activeTimers)
		default:
			break Loop
		}
	}
}

// Process immediate tasks (setImmediate phase)
func (r *Runtime) processImmediateTasks() {
	for len(r.eventLoop.immediateQueue) > 0 {
		immediateTask := r.eventLoop.immediateQueue[0]
		r.eventLoop.immediateQueue = r.eventLoop.immediateQueue[1:]
		immediateTask()
		r.processMicrotasks()
	}
}

// Process the main task queue
func (r *Runtime) processTasks() {
	for len(r.eventLoop.taskQueue) > 0 {
		task := r.eventLoop.taskQueue[0]
		r.eventLoop.taskQueue = r.eventLoop.taskQueue[1:]
		task()
		r.processMicrotasks()
	}
}

// Method to resolve the Promise with a result
func (r *Runtime) resolve(p *Promise, result interface{}) {
	// Once the promise is resolved, schedule all .then callbacks in the microtask queue
	// fmt.Printf("resolving function, num callbacks are %v\n", le(p.callbacks))
	p.result = result
	p.state = fulfilled
	for _, thenFunc := range p.callbacks {
		// fmt.Println("Scheduling a microtask")
		r.AddMicrotask(func() {
			// fmt.Printf("removing promise with result %v\n", result)
			// fmt.Println("calling a callback")
			thenFunc.onFulfilled(result)
		})
	}

	r.eventLoop.pendingPromises--
}

// Method to resolve the Promise with a result
func (r *Runtime) reject(p *Promise, result interface{}) {
	// Once the promise is resolved, schedule all .then callbacks in the microtask queue
	p.result = result
	p.state = rejected
	for _, thenFunc := range p.callbacks {
		r.AddMicrotask(func() {
			// fmt.Printf("removing promise with result %v\n", result)
			thenFunc.onRejected(result)
		})
	}

	r.eventLoop.pendingPromises--
}

// Create a new Promise
func (r *Runtime) NewPromise(executor func(resolveFunc func(interface{}), rejectFunc func(interface{}))) *Promise {
	p := &Promise{}
	executor(func(result interface{}) { r.resolve(p, result) }, func(result interface{}) { r.reject(p, result) })
	// fmt.Println("Adding promise in constructor")
	r.eventLoop.pendingPromises++
	return p
}

// Then method to chain multiple callbacks once the promise resolves
func (r *Runtime) then(p *Promise, onFulfilled func(result interface{}) interface{}, onRejected func(result interface{}) interface{}) *Promise {
	if onRejected == nil {
		onRejected = func(result interface{}) interface{} {
			return result
		}
	}

	// Create a new Promise for chaining
	newPromise := &Promise{}

	if p.state == fulfilled {
		r.AddMicrotask(func() {
			r.resolve(newPromise, onFulfilled(p.result))
		})
	} else if p.state == rejected {
		r.AddMicrotask(func() {
			r.reject(newPromise, onRejected(p.result))
		})
	} else {
		p.callbacks = append(p.callbacks, &PromiseCallback{onFulfilled: func(result interface{}) interface{} {
			// fmt.Println("resolving promise in then")
			r.resolve(newPromise, onFulfilled(result))
			return nil
		}, onRejected: func(result interface{}) interface{} {
			r.reject(newPromise, onRejected(result))
			return nil
		}})
	}

	// fmt.Println("Adding promise in then")
	r.eventLoop.pendingPromises++
	return newPromise
}

// Start the event loop, incorporating non-blocking timers
func (r *Runtime) Run() {
	r.eventLoop.running = true
	for r.eventLoop.running {

		// timer phase
		r.processTimers()
		// pending callbacks phase
		r.processTasks()
		// check phase
		r.processImmediateTasks()

		// If no more tasks and no active timers, stop the loop
		if len(r.eventLoop.taskQueue) == 0 && len(r.eventLoop.microtaskQueue) == 0 && len(r.eventLoop.immediateQueue) == 0 && r.eventLoop.activeTimers == 0 {
			if r.eventLoop.pendingPromises != 0 {
				fmt.Printf("Execution finished but %v promises still pending\n", r.eventLoop.pendingPromises)
			}
			// fmt.Println("Setting running to false")
			r.eventLoop.running = false
		}
	}
}

// func setTimeout()

func (r *Runtime) mainEventLoopFunc() interface{} {
	// Add a task to the event loop
	r.AddTask(func() {
		fmt.Println("Task 1 regular task")
	})

	// Create a new Promise that resolves after 200ms
	promise := r.NewPromise(func(resolveFunc func(interface{}), rejectFunc func(interface{})) {
		r.AddTimerTask(func() {
			fmt.Println("Timer elapsed.")
			resolveFunc("Promise resolved! in timer.")
		}, 200*time.Millisecond)
	})

	r.then(promise, func(result interface{}) interface{} {
		r.AddTask(func() {
			fmt.Printf("Task 2 regular task in Then handler 0: %v\n", result)
		})
		return nil
	}, nil)

	r.then(promise, func(result interface{}) interface{} {
		fmt.Printf("Then handler 1: %v\n", result)
		return nil
	}, nil)

	r.then(
		// Multiple .then handlers
		r.then(promise, func(result interface{}) interface{} {
			// Add a microtask to the event loop once the promise resolves
			r.AddMicrotask(func() {
				fmt.Println("Then handler 2:", result)
			})

			resultString, ok := result.(string)
			if !ok {
				return " Could not chain promise result, add err handling!"
			}
			return resultString + " Chained promise result"
		}, nil), func(result interface{}) interface{} {
			fmt.Printf("Then chained handler 3: %v\n", result)
			return nil
		}, nil)

	r.AddTask(func() {
		fmt.Println("Task 3 regular task")
	})

	r.then(promise, func(result interface{}) interface{} {
		// Add another microtask for the second then handler
		r.AddMicrotask(func() {
			fmt.Println("Then chained microtask handler 4:", result)
		})

		return nil
	}, nil)

	// Add another task
	r.AddTask(func() {
		fmt.Println("Task 4 regular task")
	})

	// Add immediate task - ran before above task
	r.AddImmediateTask(func() {
		fmt.Println("Task 5 immediate task")
	})

	return nil
}

// Start the event loop
// AddTask(eventLoop, func() {
// mainEventLoopFunc()
// })

// Run(eventLoop)

func main() {
	r := &Runtime{
		eventLoop: &EventLoop{
			timerChan:      make(chan Task), // Buffer for timer tasks
			taskQueue:      make([]Task, 0),
			microtaskQueue: make([]Task, 0),
			immediateQueue: make([]Task, 0),
		},
	}

	r.AddTask(func() {
		r.MapBoardExample()
	})

	// run the runtime!
	r.Run()
}
