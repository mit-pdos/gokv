package replica

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconfig/replica/becomeprimaryargs_gk"
	"github.com/mit-pdos/gokv/reconnectclient"
	"github.com/mit-pdos/gokv/vrsm/replica/applyasbackupargs_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/applyreply_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/err_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/getstateargs_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/getstatereply_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/increasecommitargs_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/setstateargs_gk"
)

type Clerk struct {
	cl *reconnectclient.ReconnectingClient
}

const (
	RPC_APPLYASBACKUP = uint64(0)
	RPC_SETSTATE      = uint64(1)
	RPC_GETSTATE      = uint64(2)
	RPC_BECOMEPRIMARY = uint64(3)
	RPC_PRIMARYAPPLY  = uint64(4)
	// RPC_ROAPPLYASBACKUP = uint64(5)
	RPC_ROPRIMARYAPPLY = uint64(6)
	RPC_INCREASECOMMIT = uint64(7)
)

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: reconnectclient.MakeReconnectingClient(host)}
}

func (ck *Clerk) ApplyAsBackup(args *applyasbackupargs_gk.S) err_gk.E {
	reply := new([]byte)
	err := ck.cl.Call(RPC_APPLYASBACKUP, applyasbackupargs_gk.Marshal(make([]byte, 0), *args), reply, 1000 /* ms */)
	if err != 0 {
		return err_gk.Timeout
	} else {
		e, _ := err_gk.Unmarshal(*reply)
		return e
	}
}

func (ck *Clerk) SetState(args *setstateargs_gk.S) err_gk.E {
	reply := new([]byte)
	err := ck.cl.Call(RPC_SETSTATE, setstateargs_gk.Marshal(make([]byte, 0), *args), reply, 10000 /* ms */)
	if err != 0 {
		return err_gk.Timeout
	} else {
		e, _ := err_gk.Unmarshal(*reply)
		return e
	}
}

func (ck *Clerk) GetState(args *getstateargs_gk.S) *getstatereply_gk.S {
	reply := new([]byte)
	// XXX: high timeout for this, because if the state is large, it will take a
	// long time to get.
	err := ck.cl.Call(RPC_GETSTATE, getstateargs_gk.Marshal(make([]byte, 0), *args), reply, 10000 /* ms */)
	if err != 0 {
		return &getstatereply_gk.S{Err: err_gk.Timeout}
	} else {
		rep, _ := getstatereply_gk.Unmarshal(*reply)
		return &rep
	}
}

func (ck *Clerk) BecomePrimary(args *becomeprimaryargs_gk.S) err_gk.E {
	reply := new([]byte)
	err := ck.cl.Call(RPC_BECOMEPRIMARY, becomeprimaryargs_gk.Marshal(make([]byte, 0), *args), reply, 100 /* ms */)
	if err != 0 {
		return err_gk.Timeout
	} else {
		e, _ := err_gk.Unmarshal(*reply)
		return e
	}
}

func (ck *Clerk) Apply(op []byte) (err_gk.E, []byte) {
	reply := new([]byte)
	err := ck.cl.Call(RPC_PRIMARYAPPLY, op, reply, 5000 /* ms */)
	if err == 0 {
		r, _ := applyreply_gk.Unmarshal(*reply)
		return r.Err, r.Reply
	} else {
		return err_gk.Timeout, nil
	}
}

func (ck *Clerk) ApplyRo(op []byte) (err_gk.E, []byte) {
	reply := new([]byte)
	err := ck.cl.Call(RPC_ROPRIMARYAPPLY, op, reply, 1000 /* ms */)
	if err == 0 {
		r, _ := applyreply_gk.Unmarshal(*reply)
		return r.Err, r.Reply
	} else {
		return err_gk.Timeout, nil
	}
}

func (ck *Clerk) IncreaseCommitIndex(n uint64) err_gk.E {
	return err_gk.E(ck.cl.Call(RPC_INCREASECOMMIT, increasecommitargs_gk.Marshal(make([]byte, 0), increasecommitargs_gk.S{V: n}), new([]byte), 100 /* ms */))
}
