package closed

import (
	"github.com/mit-pdos/gokv/bank"
	"github.com/mit-pdos/gokv/cachekv"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/lockservice"
	"github.com/mit-pdos/gokv/vrsm/apps/vkv"
	"github.com/mit-pdos/gokv/vrsm/configservice"
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

func mk_lconfig_paxosHosts() []grove_ffi.Address {
	var configHosts = make([]grove_ffi.Address, 0)
	return append(configHosts, lconfigHostPaxos)
}

func mk_dconfig_paxosHosts() []grove_ffi.Address {
	var configHosts = make([]grove_ffi.Address, 0)
	return append(configHosts, dconfigHostPaxos)
}

func lconfig_main(fname string) {
	var servers = make([]uint64, 0)
	servers = append(servers, lr1)
	servers = append(servers, lr2)
	configservice.StartServer(fname, lconfigHost, lconfigHostPaxos, mk_lconfig_paxosHosts(), servers)
}

func dconfig_main(fname string) {
	var servers = make([]uint64, 0)
	servers = append(servers, dr1)
	servers = append(servers, dr2)
	configservice.StartServer(fname, dconfigHost, dconfigHostPaxos, mk_dconfig_paxosHosts(), servers)
}

func kv_replica_main(fname string, me, configHost grove_ffi.Address) {
	x := new(uint64)
	*x = uint64(1)
	var configHosts = make([]grove_ffi.Address, 0)
	configHosts = append(configHosts, configHost)
	vkv.Start(fname, me, configHosts)
}

func makeBankClerk() *bank.BankClerk {
	// data server is used via cachekv; all clients must follow the cachekv protocol.
	kvck := cachekv.Make(vkv.MakeKv(mk_dconfig_hosts()))
	lck := lockservice.MakeLockClerk(vkv.MakeKv(mk_lconfig_hosts()))
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
