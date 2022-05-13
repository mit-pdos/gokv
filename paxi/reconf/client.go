package reconf

import (
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/grove_ffi"
)

type ClerkPool struct {
	cl *connman.ConnMan
}

const (
	RPC_PREPARE           = uint64(0)
	RPC_PROPOSE           = uint64(1)
	RPC_TRY_COMMIT_VAL    = uint64(1)
	RPC_TRY_CONFIG_CHANGE = uint64(1)
)

func (ck *ClerkPool) ProposeRPC(srv grove_ffi.Address, term uint64, val *MonotonicValue) bool {
	// ck.cl.CallAtLeastOnce(srv)
	return false
}

func (ck *ClerkPool) PrepareRPC(srv grove_ffi.Address, newTerm uint64, reply_ptr *PrepareReply) {
}
