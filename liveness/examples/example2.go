package examples

// Example in which both (e1 ; e2) and (e2 ; e1) fail to be live, but (e1 || e2)
// is live.
// Also a simple example which shows the problem of ciruclar reasoning
func example2() {
	x := MakeAtomicUint64()
	y := MakeAtomicUint64()

	go func() { // thread B
		for y.Read() == 0 {
		}
		x.Write(1)
	}()

	// thread A
	y.Write(1)
	for x.Read() == 0 {
	}
}

// non-live example, which ciruclar reasoning could prove live
func badexample2() {
	x := MakeAtomicUint64()
	y := MakeAtomicUint64()

	go func() { // thread B
		for y.Read() == 0 {
		}
		x.Write(1)
	}()

	// thread A
	// XXX: the for loop and the write to y here are swapped
	for x.Read() == 0 {
	}
	y.Write(1)
}
