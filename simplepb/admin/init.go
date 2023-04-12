package admin

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
)

func InitializeSystem(configHost grove_ffi.Address, servers []grove_ffi.Address) e.Error {
	configCk := config.MakeClerk(configHost)

	// Inform the config service saying the `servers` is the configuration for epoch 0
	configCk.WriteConfig(0, servers)

	// "Reconfigure" into the real next epoch, in which a servers[0] can actually
	// become primary.
	return EnterNewConfig(configHost, servers)
}
