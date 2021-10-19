package controller

import (
	"github.com/mit-pdos/gokv/pb"
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
	"time"
	"log"
	"github.com/tchajed/marshal"
)

type ControllerServer struct {
	mu       *sync.Mutex
	cn       uint64
	conf     *pb.Configuration
	hbtimers []*time.Timer
	failed   map[uint64]bool
}

func (s *ControllerServer) HeartbeatThread() {
	HBTIMEOUT := time.Duration(uint64(2)) * time.Second

	for {
		// start making heartbeats
		s.mu.Lock()
		conf := s.conf
		cn := s.cn
		s.mu.Unlock()

		hbtimers := make([]*time.Timer, len(conf.Replicas))
		clerks := make([]*pb.ReplicaClerk, len(conf.Replicas))
		for i, r := range conf.Replicas {
			// TODO: multipar this
			clerks[i] = pb.MakeReplicaClerk(r)
		}

		s.failed = make(map[uint64]bool)
		for i, _ := range clerks {
			i := i
			hbtimers[i] = time.AfterFunc(HBTIMEOUT, func() {
				s.mu.Lock()
				if s.cn == cn {
					s.failed[uint64(i)] = true
				}
				s.mu.Unlock()
			})
		}

		for {
			s.mu.Lock()
			if s.cn > cn {
				s.mu.Unlock()
				break
			}
			// NOTE: to better optimize failover, we could put this
			// HandleFailedReplicas in another thread and have it signaled by
			// the timers, rather than polling for a failed replica every X
			// seconds.
			if len(s.failed) > 0 {
				s.HandleFailedReplicas()
				s.mu.Unlock()
				break
			}
			s.mu.Unlock()

			for i, ck := range clerks {
				i := i
				ck := ck
				go func() {
					if ck.HeartbeatRPC() {
						hbtimers[i].Reset(HBTIMEOUT)
					}
				}()
			}
			time.Sleep(time.Duration(uint64(500)) * time.Millisecond)
		}
	}
}

func (s *ControllerServer) HandleFailedReplicas() {
	log.Printf("In config %d, %+v failed", s.cn, s.failed)
	n := uint64(len(s.conf.Replicas)) - uint64(len(s.failed))
	var newReplicas = make([]rpc.HostName, 0, n)
	for i, r := range s.conf.Replicas {
		if !s.failed[uint64(i)] {
			newReplicas = append(newReplicas, r)
		}
	}

	s.conf = &pb.Configuration{Replicas: newReplicas}
	s.cn += 1

	ck := pb.MakeReplicaClerk(newReplicas[0])
	ck.BecomePrimaryRPC(&pb.BecomePrimaryArgs{Cn: s.cn, Conf: s.conf})
}

func (s *ControllerServer) AddNewServerRPC(newServer rpc.HostName) {
	s.mu.Lock()
	s.cn = s.cn + 1
	s.conf = &pb.Configuration{Replicas: append(s.conf.Replicas, newServer)}
	ck := pb.MakeReplicaClerk(s.conf.Replicas[0])
	ck.BecomePrimaryRPC(&pb.BecomePrimaryArgs{Cn: s.cn, Conf: s.conf})
	s.mu.Unlock()
}

// This should be invoked locally by services to attempt appending op to the
// log
func StartControllerServer(me rpc.HostName, replicas []rpc.HostName) {
	s := new(ControllerServer)
	s.mu = new(sync.Mutex)
	s.cn = 1
	s.conf = &pb.Configuration{Replicas: replicas}

	ck := pb.MakeReplicaClerk(replicas[0])
	ck.BecomePrimaryRPC(&pb.BecomePrimaryArgs{Cn: 1, Conf: s.conf})

	go func() { s.HeartbeatThread() }() // for goose; this is silly

	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[CONTROLLER_ADD] = func(raw_args []byte, _ *[]byte) {
		dec := marshal.NewDec(raw_args)
		newServer := dec.GetInt()
		s.AddNewServerRPC(newServer)
	}
	r := rpc.MakeRPCServer(handlers)
	r.Serve(me, 1)
}
