package pb

type StateMachine struct {
	StartApply        func(op Op) ([]byte, func())
	SetStateAndUnseal func(snap []byte, nextIndex uint64, epoch uint64)
	GetStateAndSeal   func() []byte
}

type SyncStateMachine struct {
	Apply             func(op Op) []byte
	SetStateAndUnseal func(snap []byte, nextIndex uint64, epoch uint64)
	GetStateAndSeal   func() []byte
}
