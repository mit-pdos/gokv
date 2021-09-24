package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	// "github.com/tchajed/marshal"
)

type PBConfiguration struct {
	cn       uint64
	replicas []rpc.HostName
	primary  rpc.HostName
}

func EncodePBConfiguration(p *PBConfiguration) []byte {
	// FIXME: impl
	return nil
}

func DecodePBConfiguration(data []byte, p *PBConfiguration) {
	// FIXME: impl
}
