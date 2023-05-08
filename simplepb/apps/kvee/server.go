package kvee

// Replicated exactly-once KV server using simplelog for durability.

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/apps/eesm"
	"github.com/mit-pdos/gokv/simplepb/apps/kv64"
	"github.com/mit-pdos/gokv/simplepb/simplelog"
)

func Start(host grove_ffi.Address, fname string, confHost grove_ffi.Address) {
	simplelog.MakePbServer(eesm.MakeEEKVStateMachine(kv64.MakeKVStateMachine()), fname, confHost).Serve(host)
}
