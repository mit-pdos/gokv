package admin

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconfig/config"
	"github.com/mit-pdos/gokv/reconfig/example"
	"github.com/mit-pdos/gokv/reconfig/replica"
)

func EnterNewConfig(cfgHost grove_ffi.Address, servers []grove_ffi.Address) replica.Error {
	confCk := config.MakeClerk(cfgHost)
	epoch, conf_enc := confCk.GetFreshEpochAndRead()
	oldServers := replica.DecodeConfiguration(conf_enc).Replicas

	// figure out which servers are actually new
	oldServerSet := make(map[grove_ffi.Address]bool)
	for _, oldServer := range oldServers {
		oldServerSet[oldServer] = true
	}

	remainingServers := make([]grove_ffi.Address, 0)
	newServers := make([]grove_ffi.Address, 0)
	for _, server := range servers {
		if oldServerSet[server] {
			remainingServers = append(remainingServers, server)
		} else {
			newServers = append(newServers, server)
		}
	}

	// make application clerks for new servers
	newClerks := replica.FmapList(newServers, example.MakeClerk)

	// pick a replica of the old config
	oldCk := example.MakeClerk(oldServers[13%len(oldServers)])
	// get state, and tell it to stop truncating
	// FIXME: the server should be allowed to eventually truncate
	// options:
	// a.) server promises not to truncate during this epoch, so if it wants to
	// 	   truncate it has to bump epoch
	// b.) server can truncate whenever it wants, this is just an "optimization"
	index, state := oldCk.GetStateAndStopTruncation()

	// transfer state to all the new servers
	for _, newClerk := range newClerks {
		newClerk.SetState(index, state)
	}

	// LOG ALIGNMENT
	// get log
	oldLogCk := replica.MakeClerk(oldServers[13%len(oldServers)])
	err, startIndex, log := oldLogCk.GetUncommittedLog(epoch) // this is where the old config becomes unavailable
	if err != replica.ENone {
		return err
	}

	// send logs to replicas that were in prev config; guaranteed not to return EIncompleteLog
	args := &replica.BecomeReplicaArgs{Epoch: epoch, StartIndex: startIndex, Log: log}
	remainingLogClerks := replica.FmapList(remainingServers, replica.MakeClerk)

	for _, remainingClerk := range remainingLogClerks {
		err := remainingClerk.RemainReplica(args)
		if err == replica.EStale {
			return err
		}
	}

	// send logs to new servers that weren't in prev config
	success := true
	newLogClerks := replica.FmapList(newServers, replica.MakeClerk)
	for _, newClerk := range newLogClerks {
		err := newClerk.TryBecomeReplica(args)
		if err == replica.EStale {
			return err
		}
		if err == replica.EIncompleteLog {
			success = false
		}
	}
	if !success {
		panic("admin: unexpected EIncompleteLog during log alignment")
	}

	newLogClerks[0].BecomePrimary(&replica.BecomePrimaryArgs{Epoch: epoch, Conf: replica.Configuration{Replicas: newServers}})
	return replica.ENone
}
