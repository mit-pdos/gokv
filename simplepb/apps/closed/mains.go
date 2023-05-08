package closed

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/kvee"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/tchajed/goose/machine"
)

const (
	r1         = uint64(1)
	r2         = uint64(2)
	configHost = uint64(10)
)

func config_main() {
	var servers = make([]uint64, 0)
	servers = append(servers, r1)
	servers = append(servers, r2)
	config.MakeServer(servers).Serve(configHost)
}

func kv_replica_main(fname string, me grove_ffi.Address) {
	x := new(uint64)
	*x = uint64(1)
	kvee.Start(fname, me, configHost)
}

func kv_client_main(fname string, me grove_ffi.Address) {
	ck := kvee.MakeClerk(configHost)
	ck.Put(10, make([]byte, 10))
	v1 := ck.Get(10)
	machine.Assert(len(v1) == 10)
	ck.Put(10, make([]byte, 5))
	v2 := ck.Get(10)
	machine.Assert(len(v2) == 5)
}
