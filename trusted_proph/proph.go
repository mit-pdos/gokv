package trusted_proph

import (
	"github.com/goose-lang/primitive"
)

type ProphId = primitive.ProphId

func NewProph() ProphId {
	return primitive.NewProph()
}

func ResolveBytes(p ProphId, b []byte) {}
