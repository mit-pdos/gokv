package closed

import (
	"github.com/mit-pdos/gokv/bank"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/lockservice"
	"github.com/mit-pdos/gokv/simplepb/apps/kv"
	"github.com/mit-pdos/gokv/simplepb/config"
)

const (
	// data
	dr1         = uint64(1)
	dr2         = uint64(2)
	dconfigHost = uint64(10)
	// lock
	lr1         = uint64(101)
	lr2         = uint64(102)
	lconfigHost = uint64(110)
)

func lconfig_main() {
	var servers = make([]uint64, 0)
	servers = append(servers, lr1)
	servers = append(servers, lr2)
	config.MakeServer(servers).Serve(lconfigHost)
}

func dconfig_main() {
	var servers = make([]uint64, 0)
	servers = append(servers, dr1)
	servers = append(servers, dr2)
	config.MakeServer(servers).Serve(dconfigHost)
}

func kv_replica_main(fname string, me, configHost grove_ffi.Address) {
	x := new(uint64)
	*x = uint64(1)
	kv.Start(fname, me, configHost)
}

func makeBankClerk() *bank.BankClerk {
	kvck := kv.MakeKv(dconfigHost)
	lck := lockservice.MakeLockClerk(kv.MakeKv(lconfigHost))

	return bank.MakeBankClerk(lck, kvck, "init", "a1", "a2")
}

func bank_transferer_main() {
	bck := makeBankClerk()
	for {
		bck.SimpleTransfer()
	}
}

func bank_auditor_main() {
	bck := makeBankClerk()
	for {
		bck.SimpleAudit()
	}
}
