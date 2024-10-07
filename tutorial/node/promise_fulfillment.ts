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
    })

    p.promise.then((val) => {
        console.log(`First promise resolved with value ${val}`)
    })

    console.log('Resolving promise with value 8')
    p.resolve(8)

}

async function return_pending_promise_in_then() {
    // TODO These aren't conclusive, just need to look in node source code
    console.log("----------Test for returning a pending promise in .then() handler----------")
    const p = new Deferred<number>()
    const p2 = new Deferred<number>()
    p.promise.then((val) => {
        console.log(`First promise resolved with value ${val}`)
    })

    p2.promise.then((val) => {
        console.log(`p2 promise resolved with value ${val}`)
    })

    p.promise.then((val) => {
        console.log(`Returning pending promise in .then handler`)
        return p2.promise
    }).then((val) => {
        console.log(`New promise resolved with value ${val}`)
    })

    console.log('Resolving promise with value 8')
    p.resolve(8)
    console.log('Resolving p2 promise with value 5')
    
    p2.resolve(5)

}

async function jump_between_microtask_queues() {
    setImmediate(() => {
        console.log('   setImmediate 1 (check phase)');
    });
    console.log('Queued setImmediate 1');

    
}

// return_value_in_then()


// return_fulfilled_promise_in_then()

return_pending_promise_in_then()