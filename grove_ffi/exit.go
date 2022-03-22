package grove_ffi

import (
	"os"
)

func Exit(n uint64) {
	os.Exit(int(n))
}
