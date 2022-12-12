package closed

import (
	"github.com/mit-pdos/gokv/simplepb/admin"
	"github.com/mit-pdos/gokv/simplepb/apps/kv64"
	"github.com/mit-pdos/gokv/simplepb/config"
)

const (
	r1         = uint64(1)
	r2         = uint64(2)
	configHost = uint64(10)
)

func config_main() {
	config.MakeServer().Serve(configHost)
}

func kv_replica_main1() {
	x := new(uint64)
	*x = uint64(1)
	kv64.Start("kv.data", r1)
}

func kv_replica_main2() {
	kv64.Start("kv.data", r2)
}

func sysinit_main() {
	var servers = make([]uint64, 0)
	servers = append(servers, r1)
	servers = append(servers, r2)
	admin.InitializeSystem(configHost, servers)
}
