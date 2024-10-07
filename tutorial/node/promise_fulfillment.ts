import { Deferred } from "./Deferred.js";

async function return_value_in_then() {
    console.log("----------Test for returning a value in .then() handler----------")
    const p = new Deferred<number>()
    p.promise.then((val) => {
        console.log(`Returning ${val + 1} in .then handler`)
        return val + 1
    }).then((val) => {
        console.log(`New promise resolved with value ${val}`)
    })

    p.promise.then((val) => {
        console.log(`First promise resolved with value ${val}`)
    })

    console.log('Resolving promise with value 8')
    p.resolve(8)

}

async function return_fulfilled_promise_in_then() {
    console.log("----------Test for returning a fulfilled promise in .then() handler----------")
    const p = new Deferred<number>()
    const p2 = Promise.resolve(5)
    p.promise.then((val) => {
        console.log(`Returning fulfilled promise in .then handler`)
        return p2
    }).then((val) => {
        console.log(`New promise resolved with value ${val}`)
        console.log(`
        I think this means that returning a fulfilled promise doesn't synchronously resolve the promise,
        it still happens asynchronously. Because "nested promise on first promise" is printed before the line
        above.
        `)
    })

    p.promise.then((val) => {
        console.log(`First promise resolved with value ${val}`)
    }).then((val) => {
        console.log('nested promise on first promise')
    })

    console.log('Resolving promise with value 8')
    p.resolve(8)

}

async function return_pending_promise_in_then() {
    // TODO These aren't conclusive, just need to look in node source code
    console.log("----------Test for returning a pending promise in .then() handler----------")
    const p = new Deferred<number>()
    const p2 = new Deferred<number>()

    p2.promise.then((val) => {
        console.log('First .then on p2')
    })
    p.promise.then((val) => {
        console.log('First .then on p, returning p2.promise')
        return p2.promise
    }).then((val) => {
        console.log('.then on returned pending promise')
    })
    p.promise.then((val) => {
        console.log(`First promise resolved with value ${val}`)
        p2.promise.then((val) => {
            console.log(`p2 promise resolved with value ${val}`)
        })
    
        p.promise.then((val) => {
            console.log(`Returning pending promise in .then handler`)
            return p2.promise
        }).then((val) => {
            console.log(`New promise resolved with value ${val}`)
            console.log(`
            I think this means that returning a pending promise works in the expected way.
            The callbacks are added to the callbacks list of a promise as normal and evaluated
            just like any other promise, not in an out of order way when the returned promise is resolved.
            `)
        })
        console.log('Resolving p2 promise with value 5')
        
        p2.resolve(5)
    })

    p.promise.then((val) => {
        console.log("Another outer .then on p")
    })

    console.log('Resolving promise with value 8')
    p.resolve(8)

}

async function jump_between_microtask_queues() {
    // TODO this shows that between macrotasks, we can jump back and forth between next tick tasks and promise tasks
    // model this correctly in go
    setTimeout(() => {
        console.log("in first timeout call")
        Promise.resolve(5).then((val) => {
            console.log(`resolved promise with val ${val}, gonna add next tick task`)
            process.nextTick(() => {
                console.log("first call to next tick, adding promise microtask")
                Promise.resolve(4).then((val) => {
                    console.log(`resolved promise with val ${val}, gonna add next tick task and promise`)
                    process.nextTick(() => {
                        console.log("second call to next tick")
                    })

                    Promise.resolve(10).then((val) => {
                        console.log(`promise resolved with val ${val}`)
                    })
                })

                Promise.resolve(3).then((val) => {
                    console.log(`resolve promise with val ${val}`)
                })
            })

            process.nextTick(() => {
                console.log("third call to next tick")
            })
        })

        Promise.resolve(4).then((val) => {
            console.log(`Promise resolved with val ${val}`)
        })
    }, 0)

    setTimeout(() => {
        console.log("in second timeout call")
        console.log(`
        I think this shows that we can jump back and forth between promises and next tick microtasks before moving
        onto the next macrotask
        `)
    }, 0)

    Promise.resolve(9).then((val) => {
        console.log(`promise resolved with val ${val}`)
    })


}

function promise_queueMicrotask_and_nextTick() {
    setTimeout(() => {
        process.nextTick(() => {
            console.log("in next tick callback")
        })
        
        Promise.resolve().then(() => {
            console.log("in promise callback")
        })

        queueMicrotask(() => {
            console.log('in queueMicrotask callback')
        })

        process.nextTick(() => {
            console.log("in next tick callback 2")
        })
        
        Promise.resolve().then(() => {
            console.log("in promise callback 2")
        })

        queueMicrotask(() => {
            console.log('in queueMicrotask callback 2')
            console.log(`
            This shows that queueMicrotask gives callback the same precedence as promise microtasks,
            below next tick microtasks
            `)
        })
    }, 0)
}

// return_value_in_then()


// return_fulfilled_promise_in_then()

// return_pending_promise_in_then()

// jump_between_microtask_queues()

promise_queueMicrotask_and_nextTick()