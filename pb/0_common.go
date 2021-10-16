package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
)

type Configuration struct {
	Primary  rpc.HostName
	Replicas []rpc.HostName
}

func EncodePBConfiguration(p *Configuration) []byte {
	enc := marshal.NewEnc(8 + 8 + 8 * uint64(len(p.Replicas)))
	enc.PutInt(p.Primary)
	enc.PutInt(uint64(len(p.Replicas)))
	enc.PutInts(p.Replicas)
	return enc.Finish()
}

func DecodePBConfiguration(raw_conf []byte) *Configuration {
	c := new(Configuration)
	dec := marshal.NewDec(raw_conf)
	c.Primary = dec.GetInt()
	c.Replicas = dec.GetInts(dec.GetInt())
	return c
}
