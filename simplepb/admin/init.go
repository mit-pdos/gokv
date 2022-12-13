package admin

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
)

func InitializeSystem(configHost grove_ffi.Address, servers []grove_ffi.Address) e.Error {
	configCk := config.MakeClerk(configHost)
	// Get new epoch number from config service.
	// epoch, _ := configCk.GetEpochAndConfig()

	// Write to config service saying the new servers have up-to-date state.
	configCk.WriteConfig(0, servers)

	// Tell one of the servers to become primary.
	clerk := pb.MakeClerk(servers[0])
	clerk.BecomePrimary(&pb.BecomePrimaryArgs{Epoch: 0, Replicas: servers})
	return e.None
}
