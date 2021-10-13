package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	// "github.com/tchajed/marshal"
)

type PBConfiguration struct {
	replicas []rpc.HostName
	primary  rpc.HostName
}

func EncodePBConfiguration(p *PBConfiguration) []byte {
	return nil
}

func DecodePBConfiguration(d []byte) *PBConfiguration {
	// FIXME: impl
	return nil
}
