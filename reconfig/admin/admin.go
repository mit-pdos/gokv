package admin

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconfig/config"
	"github.com/mit-pdos/gokv/reconfig/example"
	"github.com/mit-pdos/gokv/reconfig/replica"
)

func EnterNewConfig(cfgHost grove_ffi.Address, newServers []grove_ffi.Address) replica.Error {
	confCk := config.MakeClerk(cfgHost)
	epoch, _ := confCk.GetFreshEpochAndRead()
	oldServers := make([]grove_ffi.Address, 0)
	if true {
		panic("admin: unmarshal configuration") // FIXME
	}

	// TODO: figure out which servers are actually new
	// remainingServers := make([]grove_ffi.Address, 0)
	newServers = make([]grove_ffi.Address, 0)
	newClerks := make([]example.Clerk, 0)

	// pick a replica of the old config
	oldCk := example.MakeClerk(oldServers[13%len(oldServers)])
	// get state, and tell it to stop truncating
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
	remainingClerks := make([]replica.Clerk, 0)
	for _, remainingClerk := range remainingClerks {
		err := remainingClerk.RemainReplica(args)
		if err == replica.EStale {
			return err
		}
	}

	// send logs to new servers that weren't in prev config
	success := true
	newLogClerks := make([]replica.Clerk, 0)
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
