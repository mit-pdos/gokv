package grove_ffi

import (
	"time"
)

type Time = uint64

func Sleep(ns uint64) {
	time.Sleep(time.Duration(ns) * time.Nanosecond)
}

func TimeNow() uint64 {
	return uint64(time.Now().UnixNano())
}
