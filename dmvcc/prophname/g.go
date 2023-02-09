package prophname

import "github.com/tchajed/goose/machine"

func Get() machine.ProphId {
	return machine.NewProph()
}
