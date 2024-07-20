package trusted_proph

import (
	"github.com/goose-lang/goose/machine"
)

type ProphId = machine.ProphId

func NewProph() ProphId {
	return machine.NewProph()
}

func ResolveBytes(p ProphId, b []byte) {}
