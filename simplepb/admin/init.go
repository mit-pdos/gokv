package admin

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config2"
	"github.com/mit-pdos/gokv/simplepb/e"
)

func InitializeSystem(configHosts []grove_ffi.Address, servers []grove_ffi.Address) e.Error {
	configCk := config2.MakeClerk(configHosts)

	// Inform the config service saying the `servers` is the configuration for epoch 0
	configCk.TryWriteConfig(0, servers)

	// "Reconfigure" into the real next epoch, in which a servers[0] can actually
	// become primary.
	return EnterNewConfig(configHosts, servers)
}
