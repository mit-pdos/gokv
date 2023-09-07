package closed

import (
	"github.com/mit-pdos/gokv/bank"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/lockservice"
	"github.com/mit-pdos/gokv/simplepb/apps/kv"
	"github.com/mit-pdos/gokv/simplepb/config2"
)

const (
	// data config replicas
	dconfigHost      = uint64(11)
	dconfigHostPaxos = uint64(12)
	// data pb replicas
	dr1 = uint64(1)
	dr2 = uint64(2)

	// lock config replicas
	lconfigHost      = uint64(111)
	lconfigHostPaxos = uint64(112)
	// lock pb replicas
	lr1 = uint64(101)
	lr2 = uint64(102)
)

func mk_lconfig_hosts() []grove_ffi.Address {
	var configHosts = make([]grove_ffi.Address, 0)
	return append(configHosts, lconfigHost)
}

func mk_dconfig_hosts() []grove_ffi.Address {
	var configHosts = make([]grove_ffi.Address, 0)
	return append(configHosts, dconfigHost)
}

func lconfig_main(fname string) {
	var servers = make([]uint64, 0)
	servers = append(servers, lr1)
	servers = append(servers, lr2)
	config2.StartServer(fname, lconfigHost, lconfigHostPaxos, mk_lconfig_hosts(), servers)
}

func dconfig_main(fname string) {
	var servers = make([]uint64, 0)
	servers = append(servers, dr1)
	servers = append(servers, dr2)
	config2.StartServer(fname, dconfigHost, dconfigHostPaxos, mk_dconfig_hosts(), servers)
}

func kv_replica_main(fname string, me, configHost grove_ffi.Address) {
	x := new(uint64)
	*x = uint64(1)
	var configHosts = make([]grove_ffi.Address, 0)
	configHosts = append(configHosts, configHost)
	kv.Start(fname, me, configHosts)
}

func makeBankClerk() *bank.BankClerk {
	kvck := kv.MakeKv(mk_dconfig_hosts())
	lck := lockservice.MakeLockClerk(kv.MakeKv(mk_lconfig_hosts()))
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
