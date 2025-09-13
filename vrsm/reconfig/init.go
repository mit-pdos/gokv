package reconfig

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/vrsm/configservice"
	"github.com/mit-pdos/gokv/vrsm/configservice/config_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/err_gk"
)

func InitializeSystem(configHosts []grove_ffi.Address, servers []grove_ffi.Address) err_gk.E {
	configCk := configservice.MakeClerk(configHosts)

	// Inform the config service saying the `servers` is the configuration for epoch 0
	configCk.TryWriteConfig(0, config_gk.S{Addrs: servers})

	// "Reconfigure" into the real next epoch, in which a servers[0] can actually
	// become primary.
	return EnterNewConfig(configHosts, servers)
}
