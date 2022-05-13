package reconf

import (
	"github.com/mit-pdos/gokv/grove_ffi"
)

type ClerkPool struct {
}

func (ck *ClerkPool) ProposeRPC(srv grove_ffi.Address, term uint64, val *MonotonicValue) bool {
	return false
}

func (ck *ClerkPool) PrepareRPC(srv grove_ffi.Address, newTerm uint64, reply_ptr *PrepareReply) {
}
