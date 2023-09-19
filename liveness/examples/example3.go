package examples

import "sync"

type nat = uint64

func example3() {
	// XXX: this example is most interesting when we have nat instead of uint64.
	// As-us, we can use the fact that n < 2**64 to bound the number of steps it
	// takes to terminate.
	N := new(nat)
	mu := new(sync.Mutex)

	go func() { // thread B
		for {
			mu.Lock()
			*N++
			mu.Unlock()
		}
	}()

	// thread A
	mu.Lock()
	n := *N
	mu.Unlock()

	i := nat(0)
	for i < n {
		i++
	}
	return
}
