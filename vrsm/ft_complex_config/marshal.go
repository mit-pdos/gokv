package ftconfig

import "github.com/mit-pdos/gokv/vrsm/e"

type ApplyAsFollowerArgs struct {
	state     []byte
	epoch     Epoch
	nextIndex uint64
}

type BecomeFollowerArgs struct {
	epoch Epoch
}

type BecomeFollowerReply struct {
	// if no error, then epoch = acceptedEpoch for the given state, with
	// acceptedEpoch <= args.epoch
	// if error = e.Stale, then epoch = epoch of the follower server, which
	// should be greater than or equal to args.epoch.
	err   e.Error
	epoch Epoch

	nextIndex  uint64
	state      []byte
	nextConfig []uint64
	config     []uint64
}
