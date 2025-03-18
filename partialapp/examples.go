package main

func partiallyApplyMe(x string, y int) {
	if len(x) != y {
		panic("not allowed")
	}
}

type Foo string

func (f Foo) someMethod() {
}

func (f Foo) someMethodWithArgs(y string, z int) {
	partiallyApplyMe(string(f)+y, z)
}

func main() {
	var x func(x string, y int)
	x = partiallyApplyMe
	x("blah", 4)

	defer x("ok", 2)
	defer partiallyApplyMe("abc", 3)

	f := Foo("a")
	f.someMethod()

	x = f.someMethodWithArgs
	x("b", 2)
	f.someMethodWithArgs("bc", 3)
}
