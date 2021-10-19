package time

import (
	"time"
)

type Timer struct {
	t *time.Timer
}

const Millisecond = uint64(time.Millisecond)
const Second = uint64(time.Second)

func AfterFunc(duration uint64, f func()) *Timer {
	return &Timer{time.AfterFunc(time.Duration(duration) * time.Nanosecond, f)}
}

func (t *Timer) Reset(duration uint64) {
	t.t.Reset(time.Duration(duration) * time.Nanosecond)
}

func Sleep(duration uint64) {
	time.Sleep(time.Duration(duration) * time.Nanosecond)
}
