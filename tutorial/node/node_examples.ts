import * as fs from 'fs';

/**
 * 
 * To run this example in the terminal:
 *  - have npm installed: https://docs.npmjs.com/downloading-and-installing-node-js-and-npm or https://nodejs.org/en/download/package-manager
 *  - have typescript installed: https://www.typescriptlang.org/download
 *  - run ./run-ts-file node_examples (no filename extension)
 * https://stackoverflow.com/questions/76035802/priority-order-between-nexttick-and-promise-in-nodejs
 */
function main(summary: string) {
    console.log('----------In main----------');
    setImmediate(() => {
        console.log('   setImmediate 1 (check phase)');
    });
    console.log('Queued setImmediate 1');

    setImmediate(() => {
        console.log('   setImmediate 2 (check phase)');
        console.log('----------Done processing callbacks, queues empty----------');
        console.log(summary);
    });
    console.log('Queued setImmediate 2');

    fs.readFile('fake_file.txt', (err, data) => {
        console.log(`   readFile I/O 1 (poll phase). Err: ${err}, Data: ${data}`);
    });
    console.log('Queued readFile 1');

    fs.readFile('fake_file.text', (err, data) => {
        console.log(`   readFile I/O 2 (poll phase). Err: ${err}, Data: ${data}`);
    });
    console.log('Queued readFile 2');

    setTimeout(() => {
        console.log('   setTimeout 1 (timers phase)');
        Promise.resolve().then(() => {
            console.log('       Promise 3 resolved (microtask) scheduled in setTimeout 1');
            process.nextTick(() => {
                console.log('           process.nextTick 3 (microtask) scheduled in promise 3 ');
            });
            console.log('       Queued netTick 3 in promise 3 ');
            process.nextTick(() => {
                console.log('           process.nextTick 4 (microtask) scheduled in promise 3 ');
            });
            console.log('       Queued netTick 4 in promise 3 ');
            Promise.resolve().then(() => {
                console.log('           Promise 4 resolved (microtask) scheduled in promise 3 ');
            });
            console.log('       Queued promise 4 in promise 3 ');
            Promise.resolve().then(() => {
                console.log('           Promise 5 resolved (microtask) scheduled in promise 3 ');
            });
            console.log('       Queued promise 5 in promise 3 ');
        })
        console.log('   Queued promise 3 in setTimeout 1');
      }, 0);
    console.log('Queued setTimeout 1');

    setTimeout(() => {
        console.log('   setTimeout 2 (timers phase)');
    }, 0);
    console.log('Queued setTimeout 2');

    Promise.resolve().then(() => {
        console.log('   Promise 1 resolved (microtask)');
    });
    console.log('Queued promise 1');

    Promise.resolve().then(() => {
        console.log('   Promise 2 resolved (microtask)');
    });
    console.log('Queued promise 2');

    process.nextTick(() => {
        console.log('----------Stack empty, processing queues----------');
        console.log('   process.nextTick 1 (microtask)');
    });
    console.log('Queued nextTick 1');

    process.nextTick(() => {
        console.log('   process.nextTick 2 (microtask)');
    });
    console.log('Queued nextTick 2');
    console.log('----------Done in main----------');
}

console.log(`
Event Loop Phases (macrotask queues):
                                
   ┌───────────────────────────┐
┌─>│           timers          │
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
│  │     pending callbacks     │
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
│  │       idle, prepare       │
│  └─────────────┬─────────────┘      ┌───────────────┐
│  ┌─────────────┴─────────────┐      │   incoming:   │
│  │           poll            │<─────┤  connections, │
│  └─────────────┬─────────────┘      │   data, etc.  │
│  ┌─────────────┴─────────────┐      └───────────────┘
│  │           check           │
│  └─────────────┬─────────────┘
│  ┌─────────────┴─────────────┐
└──┤      close callbacks      │
   └───────────────────────────┘

Example:
In the following example, we:
    - queue 2 callbacks for the check phase with setImmediate()
    - queue 2 callbacks for the poll phase with fs.readFile(). A fake file is used to cause the callback to be placed on the queue as fast as possible
    - queue 2 callbacks for the timers phase with setTimeout(). A 0 timeout is used to cause the callbacks to be placed on the queue as fast as possible. In the callback we:
        - queue 1 callback for the promises microtask queue with Promise.resolve(). In the callback, we:
            - queue 2 callbacks for the next tick microtask queue with process.nextTick().
            - queue 2 callbacks for the promises microtask queue with Promise.resolve().
    - queue 2 callbacks for the promises microtask queue with Promise.resolve(). Since they are immediately resolved, they are immediately placed on the promises microtask queue
    - queue 2 callbacks for the next tick microtask queue with process.nextTick().
`);
main(`
In summary, we see that:
    - the setImmediate(), fs.readFile(), and setTimeout() callbacks occurred in the reverse order that they were queued in,
    which confirms the order of those phases in the above diagram.
    - macrotask queues are completely flushed before continuing to the next phase of the event loop
    - between each 'tick' of the event loop, or processing of one callback of any of the macrotask queues in any of the phases,
    the promise and next tick microtasks queues are completely flushed, including any microtask callbacks appended during the processing
    of another microtask callbacks.
    - The next tick microtask queue is flushed before the promise microtask queue, but once the promises queue begins to be flushed,
    it won't be interrupted by the next tick queue.
`);

