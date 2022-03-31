package example

import (
	// "github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}
