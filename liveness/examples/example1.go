package examples

// simplest example in which thread A depends on something being done by thread
// B.

func example1() uint64 {
	x := MakeAtomicUint64()

	go func() { // thread B
		x.Write(1)
	}()

	// thread A
	for {
		if x.Read() != 0 {
			break
		}
	}
	return x.Read()
}
