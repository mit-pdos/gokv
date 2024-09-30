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

func WrapperFunc() {

	var eventLoop = &EventLoop{
		timerChan:      make(chan Task, 10), // Buffer for timer tasks
		taskQueue:      make([]Task, 0),
		microtaskQueue: make([]Task, 0),
		immediateQueue: make([]Task, 0),
	}

	// Add a task (e.g., regular async tasks)
	AddTask := func(el *EventLoop, task Task) {
		el.taskQueue = append(el.taskQueue, task)
	}

	// Add a microtask (e.g., process.nextTick, promises)
	AddMicrotask := func(el *EventLoop, microtask Task) {
		el.microtaskQueue = append(el.microtaskQueue, microtask)
	}

	// Add a task to be run in the immediate phase (setImmediate)
	AddImmediateTask := func(el *EventLoop, immediate Task) {
		el.immediateQueue = append(el.immediateQueue, immediate)
	}

	// Non-blocking AddTimerTask using goroutines and a timer channel
	AddTimerTask := func(el *EventLoop, task Task, duration time.Duration) {
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
	processMicrotasks := func(el *EventLoop) {
		for len(el.microtaskQueue) > 0 {
			microtask := el.microtaskQueue[0]
			el.microtaskQueue = el.microtaskQueue[1:]
			microtask()
		}
	}

	// Process immediate tasks (setImmediate phase)
	processImmediateTask := func(el *EventLoop) {
		if len(el.immediateQueue) == 0 {
			return
		}
		immediateTask := el.immediateQueue[0]
		el.immediateQueue = el.immediateQueue[1:]
		immediateTask()
	}

	// Process the main task queue
	processTask := func(el *EventLoop) {
		if len(el.taskQueue) == 0 {
			return
		}

		task := el.taskQueue[0]
		el.taskQueue = el.taskQueue[1:]
		task()
	}

	// Method to resolve the Promise with a result
	resolve := func(p *Promise, result interface{}) {
		// Once the promise is resolved, schedule all .then callbacks in the microtask queue
		// fmt.Printf("resolving function, num callbacks are %v\n", len(p.callbacks))
		p.result = result
		p.state = fulfilled
		for _, thenFunc := range p.callbacks {
			// fmt.Println("Scheduling a microtask")
			AddMicrotask(eventLoop, func() {
				// fmt.Printf("removing promise with result %v\n", result)
				// fmt.Println("calling a callback")
				thenFunc.onFulfilled(result)
			})
		}

		eventLoop.pendingPromises--
	}

	// Method to resolve the Promise with a result
	reject := func(p *Promise, result interface{}) {
		// Once the promise is resolved, schedule all .then callbacks in the microtask queue
		p.result = result
		p.state = rejected
		for _, thenFunc := range p.callbacks {
			AddMicrotask(eventLoop, func() {
				// fmt.Printf("removing promise with result %v\n", result)
				thenFunc.onRejected(result)
			})
		}

		eventLoop.pendingPromises--
	}

	// Create a new Promise
	NewPromise := func(executor func(resolveFunc func(interface{}), rejectFunc func(interface{}))) *Promise {
		p := &Promise{}
		executor(func(result interface{}) { resolve(p, result) }, func(result interface{}) { reject(p, result) })
		// fmt.Println("Adding promise in constructor")
		eventLoop.pendingPromises++
		return p
	}

	// Then method to chain multiple callbacks once the promise resolves
	then := func(p *Promise, onFulfilled func(result interface{}) interface{}, onRejected func(result interface{}) interface{}) *Promise {
		if onRejected == nil {
			onRejected = func(result interface{}) interface{} {
				return result
			}
		}

		// Create a new Promise for chaining
		newPromise := &Promise{}

		if p.state == fulfilled {
			AddMicrotask(eventLoop, func() {
				resolve(newPromise, onFulfilled(p.result))
			})
		} else if p.state == rejected {
			AddMicrotask(eventLoop, func() {
				reject(newPromise, onRejected(p.result))
			})
		} else {
			p.callbacks = append(p.callbacks, &PromiseCallback{onFulfilled: func(result interface{}) interface{} {
				// fmt.Println("resolving promise in then")
				resolve(newPromise, onFulfilled(result))
				return nil
			}, onRejected: func(result interface{}) interface{} {
				reject(newPromise, onRejected(result))
				return nil
			}})
		}

		// fmt.Println("Adding promise in then")
		eventLoop.pendingPromises++
		return newPromise
	}

	// Start the event loop, incorporating non-blocking timers
	Run := func(el *EventLoop) {
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
				processMicrotasks(el)
				processTask(el)
				processImmediateTask(el)
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

	mainEventLoopFunc := func() interface{} {
		// Add a task to the event loop
		AddTask(eventLoop, func() {
			fmt.Println("Task 1 regular task")
		})

		// Create a new Promise that resolves after 200ms
		promise := NewPromise(func(resolveFunc func(interface{}), rejectFunc func(interface{})) {
			AddTimerTask(eventLoop, func() {
				resolveFunc("Promise resolved! in timer.")
			}, 200*time.Millisecond)
		})

		then(promise, func(result interface{}) interface{} {
			AddTask(eventLoop, func() {
				fmt.Printf("Task 2 regular task in Then handler 0: %v\n", result)
			})
			return nil
		}, nil)

		then(promise, func(result interface{}) interface{} {
			fmt.Printf("Then handler 1: %v\n", result)
			return nil
		}, nil)

		then(
			// Multiple .then handlers
			then(promise, func(result interface{}) interface{} {
				// Add a microtask to the event loop once the promise resolves
				AddMicrotask(eventLoop, func() {
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

		AddTask(eventLoop, func() {
			fmt.Println("Task 3 regular task")
		})

		then(promise, func(result interface{}) interface{} {
			// Add another microtask for the second then handler
			AddMicrotask(eventLoop, func() {
				fmt.Println("Then chained microtask handler 4:", result)
			})

			return nil
		}, nil)

		// Add another task
		AddTask(eventLoop, func() {
			fmt.Println("Task 4 regular task")
		})

		// Add immediate task - ran before above task
		AddImmediateTask(eventLoop, func() {
			fmt.Println("Task 5 immediate task")
		})

		return nil
	}

	// Start the event loop
	// AddTask(eventLoop, func() {
	mainEventLoopFunc()
	// })

	Run(eventLoop)

}

func main() {
	WrapperFunc()
}
