package ftconfig

import "github.com/mit-pdos/gokv/simplepb/e"

type FollowerClerk struct {
}

func (ck *FollowerClerk) BecomeFollower(args *BecomeFollowerArgs, reply *BecomeFollowerReply) {
	panic("impl")
}

func (ck *FollowerClerk) GetEpoch() (e.Error, Epoch) {
	panic("impl")
}
