package examples

import (
	"sync"
)

func example4(nondet bool) {
	mu := new(sync.Mutex)
	var request bool = false
	request_cond := sync.NewCond(mu)

	x := MakeAtomicUint64()

	// thread B
	go func() {
		mu.Lock()
		for {
			for !request {
				request_cond.Wait()
			}
			x.Write(10)
		}
	}()

	// thread A
	if nondet {
		request = true
		request_cond.Signal()

		for x.Read() == 0 {
		}
	}
}
