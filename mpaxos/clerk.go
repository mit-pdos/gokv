package mpaxos

import (
)

// these clerks hide connection failures, and retry forever
type Clerk struct {
}

func (s *Clerk) enterNewEpoch(args *enterNewEpochArgs, reply *enterNewEpochReply) {
}

func (s *Clerk) applyAsFollower(args *applyAsFollowerArgs, reply *applyAsFollowerReply) {
}
