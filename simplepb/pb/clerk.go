package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

const (
	RPC_APPLYASBACKUP = uint64(0)
	RPC_SETSTATE      = uint64(1)
	RPC_GETSTATE      = uint64(2)
	RPC_BECOMEPRIMARY = uint64(3)
	RPC_PRIMARYAPPLY  = uint64(4)
)

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}

func (ck *Clerk) ApplyAsBackup(args *ApplyAsBackupArgs) e.Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_APPLYASBACKUP, EncodeApplyAsBackupArgs(args), reply, 100 /* ms */)
	if err != 0 {
		return e.Timeout
	} else {
		return e.DecodeError(*reply)
	}
}

func (ck *Clerk) SetState(args *SetStateArgs) e.Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_SETSTATE, EncodeSetStateArgs(args), reply, 1000 /* ms */)
	if err != 0 {
		return e.Timeout
	} else {
		return e.DecodeError(*reply)
	}
}

func (ck *Clerk) GetState(args *GetStateArgs) *GetStateReply {
	reply := new([]byte)
	err := ck.cl.Call(RPC_GETSTATE, EncodeGetStateArgs(args), reply, 1000 /* ms */)
	if err != 0 {
		return &GetStateReply{Err: e.Timeout}
	} else {
		return DecodeGetStateReply(*reply)
	}
}

func (ck *Clerk) BecomePrimary(args *BecomePrimaryArgs) e.Error {
	reply := new([]byte)
	err := ck.cl.Call(RPC_BECOMEPRIMARY, EncodeBecomePrimaryArgs(args), reply, 100 /* ms */)
	if err != 0 {
		return e.Timeout
	} else {
		return e.DecodeError(*reply)
	}
}

func (ck *Clerk) Apply(op []byte) (e.Error, []byte) {
	reply := new([]byte)
	err := ck.cl.Call(RPC_PRIMARYAPPLY, op, reply, 2000 /* ms */)
	if err == 0 {
		r := DecodeApplyReply(*reply)
		return r.Err, r.Reply
	} else {
		return err, nil
	}
}
