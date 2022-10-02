// state-transfer reconfiguration

package st_reconfig

/*
import (
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
)

type ExtraFunctions struct {
	SetStateAndEpochAndUnseal func([]byte, uint64, uint64)
	GetStateAndSeal           func() []byte
}

type Server struct {
	sm  *pb.StateMachine
	esm *ExtraFunctions
}

func (s *Server) SetState(args *SetStateArgs) e.Error {
	s.sm.Mu.Lock()
	if s.sm.GetEpoch() > args.Epoch {
		s.sm.Mu.Unlock()
		return e.Stale
	} else if s.sm.GetEpoch() == args.Epoch {
		s.sm.Mu.Unlock()
		return e.None
	} else {
		// FIXME: have to call into base server
		s.isPrimary = false

		s.esm.SetStateAndEpochAndUnseal(args.State, args.Epoch, args.NextIndex)

		s.sm.Mu.Unlock()
		return e.None
	}
}

// XXX: probably should rename to GetStateAndSeal
func (s *Server) GetState(args *GetStateArgs) *GetStateReply {
	s.sm.Mu.Lock()
	if s.isEpochStale(args.Epoch) {
		s.sm.Mu.Unlock()
		return &GetStateReply{Err: e.Stale, State: nil}
	}

	ret := s.esm.GetStateAndSeal()
	nextIndex := s.sm.GetNextIndex()
	s.sm.Mu.Unlock()

	return &GetStateReply{Err: e.None, State: ret, NextIndex: nextIndex}
}
*/
