package controller

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/mit-pdos/gokv/pb"
	"sync"
)

type ControllerServer struct {
	mu   *sync.Mutex
	cn   uint64
	conf *pb.Configuration
}

func (s *ControllerServer) HeartbeatThread() {
	// FIXME: impl
}

// This should be invoked locally by services to attempt appending op to the
// log
func  StartControllerServer(me rpc.HostName, primary rpc.HostName, replicas []rpc.HostName) {
	s := new(ControllerServer)
	s.mu = new(sync.Mutex)
	s.conf = &pb.Configuration{Primary:primary, Replicas:replicas}

	ck := pb.MakeReplicaClerk(primary)
	ck.BecomePrimaryRPC(&pb.BecomePrimaryArgs{Cn:1, Conf:s.conf})
}
