package exactlyonce

type VersionedStateMachine struct {
	ApplyVolatile func(op []byte, vnum uint64) []byte
	ApplyReadonly func(op []byte) (uint64, []byte)
	SetState      func(snap []byte, nextIndex uint64)
	GetState      func() []byte
}
