package closed

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/kv"
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
	kv.Start(fname, me, configHost)
}

func kv_client_main(fname string, me grove_ffi.Address) {
	ck := kv.MakeClerk(configHost)
	ck.Put("a", "ABCD")
	v1 := ck.Get("a")
	machine.Assert(len(v1) == 4)
	ck.Put("a", "EFG")
	v2 := ck.Get("a")
	machine.Assert(len(v2) == 3)
}
