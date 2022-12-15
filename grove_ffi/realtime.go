package grove_ffi

import (
	"time"
)

func GetTimeRange() (uint64, uint64) {
	// FIXME: implement this correctly
	t := uint64(time.Now().UnixNano())
	return t, t + 10e6
}
