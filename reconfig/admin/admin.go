package admin

import (
	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconfig/config"
	"github.com/mit-pdos/gokv/reconfig/example"
	"github.com/mit-pdos/gokv/reconfig/replica"
	"github.com/mit-pdos/gokv/reconfig/replica/becomeprimaryargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/becomereplicaargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/configuration_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/error_gk"
)

func EnterNewConfig(cfgHost grove_ffi.Address, servers []grove_ffi.Address) error_gk.E {
	confCk := config.MakeClerk(cfgHost)
	epoch, conf_enc := confCk.GetFreshEpochAndRead()
	conf, _ := configuration_gk.Unmarshal(conf_enc)
	oldServers := conf.Replicas

	// figure out which servers are actually new
	oldServerSet := make(map[grove_ffi.Address]bool)
	for _, oldServer := range oldServers {
		oldServerSet[oldServer] = true
	}

	var remainingServers = make([]grove_ffi.Address, 0)
	var newServers = make([]grove_ffi.Address, 0)
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
	if err != error_gk.ENone {
		return err
	}

	// send logs to replicas that were in prev config; guaranteed not to return EIncompleteLog
	args := &becomereplicaargs_gk.S{Epoch: epoch, StartIndex: startIndex, Log: log}
	remainingLogClerks := replica.FmapList(remainingServers, replica.MakeClerk)

	for _, remainingClerk := range remainingLogClerks {
		err := remainingClerk.RemainReplica(args)
		if err == error_gk.EStale {
			return err
		}
	}

	// send logs to new servers that weren't in prev config; theoretically can
	// return EIncompleteLog, if the state snapshot is too old.
	success := true
	newLogClerks := replica.FmapList(newServers, replica.MakeClerk)
	for _, newClerk := range newLogClerks {
		err := newClerk.TryBecomeReplica(args)
		if err == error_gk.EStale {
			return err
		}
		if err == error_gk.EIncompleteLog {
			success = false
		}
	}
	if !success {
		panic("admin: unexpected EIncompleteLog during log alignment")
	}

	newLogClerks[0].BecomePrimary(&becomeprimaryargs_gk.S{Epoch: epoch, Conf: configuration_gk.S{Replicas: newServers}})
	return error_gk.ENone
}

// A better protocol
func EnterNewConfig2(cfgHost grove_ffi.Address, servers []grove_ffi.Address) error_gk.E {
	var err = error_gk.ENone
	confCk := config.MakeClerk(cfgHost)
	epoch, conf_enc := confCk.GetFreshEpochAndRead()
	conf, _ := configuration_gk.Unmarshal(conf_enc)
	oldServers := conf.Replicas

	// figure out which servers are actually new
	oldServerSet := make(map[grove_ffi.Address]bool)
	for _, oldServer := range oldServers {
		oldServerSet[oldServer] = true
	}

	var remainingServers = make([]grove_ffi.Address, 0)
	var newServers = make([]grove_ffi.Address, 0)
	for _, server := range servers {
		if oldServerSet[server] {
			remainingServers = append(remainingServers, server)
		} else {
			newServers = append(newServers, server)
		}
	}

	// get log
	oldLogCk := replica.MakeClerk(oldServers[primitive.RandomUint64()%uint64(len(oldServers))])
	err, startIndex, log := oldLogCk.GetUncommittedLog(epoch)
	if err != error_gk.ENone {
		return err
	}

	// STEP 1
	// add servers to membership for the log replication; log alignment to make
	// sure the new servers have the same entries as old log servers.

	// send logs to replicas that were in prev config; guaranteed not to return EIncompleteLog
	args := &becomereplicaargs_gk.S{Epoch: epoch, StartIndex: startIndex, Log: log}
	remainingLogClerks := replica.FmapList(remainingServers, replica.MakeClerk)
	for _, remainingClerk := range remainingLogClerks {
		err = remainingClerk.RemainReplica(args)
		if err == error_gk.EStale {
			break
		} else {
			continue
		}
	}
	if err != error_gk.ENone {
		return err
	}

	// send logs to new replicas;
	newLogClerks := replica.FmapList(newServers, replica.MakeClerk)
	for _, newClerk := range newLogClerks {
		err = newClerk.TryBecomeReplica(args)
		if err == error_gk.EStale {
			break
		}
	}
	if err != error_gk.ENone {
		return err
	}

	// STEP 2
	// transfer state snapshot

	// STEP 3
	// kick out (oldServers ∖ servers)
	return err
}
