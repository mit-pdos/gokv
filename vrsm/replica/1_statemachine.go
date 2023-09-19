package replica

type StateMachine struct {
	StartApply        func(op Op) ([]byte, func())
	ApplyReadonly     func(op Op) (uint64, []byte)
	SetStateAndUnseal func(snap []byte, nextIndex uint64, epoch uint64)
	GetStateAndSeal   func() []byte
}

type SyncStateMachine struct {
	Apply             func(op Op) []byte
	ApplyReadonly     func(op Op) (uint64, []byte)
	SetStateAndUnseal func(snap []byte, nextIndex uint64, epoch uint64)
	GetStateAndSeal   func() []byte
}
