package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

type Configuration struct {
	Replicas []grove_ffi.Address
}

func EncodePBConfiguration(p *Configuration) []byte {
	enc := marshal.NewEnc(8 + 8 + 8*uint64(len(p.Replicas)))
	enc.PutInt(uint64(len(p.Replicas)))
	enc.PutInts(p.Replicas)
	return enc.Finish()
}

func DecodePBConfiguration(raw_conf []byte) *Configuration {
	c := new(Configuration)
	dec := marshal.NewDec(raw_conf)
	c.Replicas = dec.GetInts(dec.GetInt())
	return c
}
