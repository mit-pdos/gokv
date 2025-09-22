package replica

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconfig/replica/appendargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/becomeprimaryargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/becomereplicaargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/error_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/logentry_gk"
	"github.com/mit-pdos/gokv/urpc"
)

type Clerk struct {
	cl *urpc.Client
}

func (ck *Clerk) appendRPC(args *appendargs_gk.S) error_gk.E {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) BecomePrimary(args *becomeprimaryargs_gk.S) error_gk.E {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) TryBecomeReplica(args *becomereplicaargs_gk.S) error_gk.E {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) RemainReplica(args *becomereplicaargs_gk.S) error_gk.E {
	// FIXME: impl
	panic("replica: impl")
}

func (ck *Clerk) GetUncommittedLog(epoch uint64) (error_gk.E, uint64, []logentry_gk.S) {
	panic("replica: impl")
}

func MakeClerk(host grove_ffi.Address) *Clerk {
	return &Clerk{cl: urpc.MakeClient(host)}
}
