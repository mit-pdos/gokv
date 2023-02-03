package atomic_commit

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
)

type Decision = byte

const (
	Unknown byte = 0
	Commit  byte = 1
	Abort   byte = 2
)

type ParticipantServer struct {
	m          *sync.Mutex
	preference bool
	// decision   Decision (must talk to coordinator)
}

func (s *ParticipantServer) GetPreference() bool {
	s.m.Lock()
	pref := s.preference
	s.m.Unlock()
	return pref
}

func MakeParticipant(pref bool) *ParticipantServer {
	return &ParticipantServer{
		m:          new(sync.Mutex),
		preference: pref,
	}
}

type ParticipantClerk struct {
	client *urpc.Client
}

type CoordinatorServer struct {
	m            *sync.Mutex
	decision     Decision
	preferences  []Decision
	participants []*ParticipantClerk
}

type CoordinatorClerk struct {
	client *urpc.Client
}

const (
	Yes bool = true
	No  bool = false
)

const GetPreferenceId uint64 = 0

func (ck *ParticipantClerk) GetPreference() bool {
	req := make([]byte, 0)
	var reply = make([]byte, 1)
	err := ck.client.Call(GetPreferenceId, req, &reply, 1000)
	machine.Assume(err == 0) // no timeout or disconnect from participant
	b := reply[0]
	if b == 0 {
		return false
	} else {
		return true
	}
}

// make a decision once we have all the preferences
//
// assumes we have all preferences (ie, no Unknown)
func (s *CoordinatorServer) makeDecision() {
	s.m.Lock()
	for _, pref := range s.preferences {
		if pref == Abort {
			s.decision = Abort
		}
		// assert(pref != Unknown)
	}
	if s.decision == Unknown {
		s.decision = Commit
	}
	s.m.Unlock()
}

func (s *CoordinatorServer) backgroundLoop() {
	for i, h := range s.participants {
		pref := h.GetPreference()
		var decision Decision
		if pref {
			decision = Commit
		} else {
			decision = Abort
		}
		s.m.Lock()
		s.preferences[i] = decision
		s.m.Unlock()
	}
	s.makeDecision()
}

func MakeCoordinator(participants []grove_ffi.Address) *CoordinatorServer {
	decision := Unknown

	m := new(sync.Mutex)
	preferences := make([]Decision, len(participants))
	var clerks = make([]*ParticipantClerk, 0)
	for _, a := range participants {
		client := urpc.MakeClient(a)
		clerks = append(clerks, &ParticipantClerk{
			client: client,
		})
	}
	return &CoordinatorServer{
		m:            m,
		decision:     decision,
		preferences:  preferences,
		participants: clerks,
	}
}

func (ck *CoordinatorClerk) GetDecision() Decision {
	req := make([]byte, 0)
	var reply = make([]byte, 1)
	err := ck.client.Call(GetDecisionId, req, &reply, 1000)
	machine.Assume(err == 0) // no timeout or disconnect from participant
	return reply[0]
}

func (s *CoordinatorServer) GetDecision() Decision {
	s.m.Lock()
	decision := s.decision
	s.m.Unlock()
	return decision
}

const GetDecisionId uint64 = 1

func CoordinatorMain(me grove_ffi.Address, participants []grove_ffi.Address) {
	coordinator := MakeCoordinator(participants)
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[GetDecisionId] = func(_req []byte, reply *[]byte) {
		decision := coordinator.GetDecision()
		replyData := make([]byte, 1)
		replyData[0] = decision
		*reply = replyData
	}
	server := urpc.MakeServer(handlers)
	server.Serve(me)
	go func() {
		coordinator.backgroundLoop()
	}()
}

func ParticipantMain(me grove_ffi.Address, pref bool) {
	participant := MakeParticipant(pref)
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[GetPreferenceId] = func(_req []byte, reply *[]byte) {
		pref := participant.GetPreference()
		replyData := make([]byte, 1)
		if pref {
			replyData[0] = 1
		} else {
			replyData[0] = 0
		}
		*reply = replyData
	}
	server := urpc.MakeServer(handlers)
	server.Serve(me)
}
