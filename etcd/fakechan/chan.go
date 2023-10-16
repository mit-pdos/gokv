package fakechan

type Chan[T any] struct {
}

func Make() {
	panic("axiom")
}

func (c Chan[T]) Put(a T) {
	panic("axiom")
}

func (c Chan[T]) Get() T {
	panic("axiom")
}

func (c Chan[T]) ForRange(f func(x T)) {
	panic("axiom")
}
