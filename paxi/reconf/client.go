package reconf

import (
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

type ClerkPool struct {
	cl *connman.ConnMan
}

const (
	RPC_PREPARE           = uint64(0)
	RPC_PROPOSE           = uint64(1)
	RPC_TRY_COMMIT_VAL    = uint64(2)
	RPC_TRY_CONFIG_CHANGE = uint64(3)
)

func MakeClerkPool() *ClerkPool {
	return &ClerkPool{cl: connman.MakeConnMan()}
}

func (ck *ClerkPool) PrepareRPC(srv grove_ffi.Address, newTerm uint64, reply_ptr *PrepareReply) {
	raw_reply := new([]byte)

	// FIXME: this should be allowed to give up, rather than loop forever
	ck.cl.CallAtLeastOnce(srv, RPC_PREPARE, marshal.WriteInt(make([]byte, 0), newTerm), raw_reply, 10 /* ms */)
	*reply_ptr = *DecPrepareReply(*raw_reply)
}

func (ck *ClerkPool) ProposeRPC(srv grove_ffi.Address, term uint64, val *MonotonicValue) bool {
	args := &ProposeArgs{Term: term, Val: val}
	raw_reply := new([]byte)

	// FIXME: this should be allowed to give up, rather than loop forever
	ck.cl.CallAtLeastOnce(srv, RPC_PROPOSE, EncProposeArgs(args), raw_reply, 10 /* ms */)
	err, _ := marshal.ReadInt(*raw_reply)
	return err == 0
}

func (ck *ClerkPool) TryCommitVal(srv grove_ffi.Address, v []byte) bool {
	raw_reply := new([]byte)

	// FIXME: this should be allowed to give up, rather than loop forever
	ck.cl.CallAtLeastOnce(srv, RPC_TRY_COMMIT_VAL, v, raw_reply, 1000 /* ms */)
	err, _ := marshal.ReadInt(*raw_reply)
	return err == 0
}

func (ck *ClerkPool) TryConfigChange(srv grove_ffi.Address, newMembers []grove_ffi.Address) bool {
	raw_args := EncMembers(newMembers)
	raw_reply := new([]byte)

	// FIXME: this should be allowed to give up, rather than loop forever
	ck.cl.CallAtLeastOnce(srv, RPC_TRY_CONFIG_CHANGE, raw_args, raw_reply, 50 /* ms */)
	err, _ := marshal.ReadInt(*raw_reply)
	return err == 0
}
