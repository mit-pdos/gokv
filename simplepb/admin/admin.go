package admin

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/goose/machine"
	"sync"
)

func EnterNewConfig(configHost grove_ffi.Address, servers []grove_ffi.Address) e.Error {
	configCk := config.MakeClerk(configHost)
	// Get new epoch number from config service.
	// Read from config service, fenced with that epoch.
	epoch, oldServers := configCk.GetEpochAndConfig()

	// Enter new epoch on one of the old servers.
	// Get a copy of the state from that old server.
	oldClerk := pb.MakeClerk(oldServers[machine.RandomUint64()%uint64(len(oldServers))])
	reply := oldClerk.GetState(&pb.GetStateArgs{Epoch: epoch})
	if reply.Err != e.None {
		return reply.Err
	}

	// Set the state of all the new servers.
	clerks := make([]*pb.Clerk, len(servers))
	for i := range clerks {
		clerks[i] = pb.MakeClerk(servers[i])
	}
	wg := new(sync.WaitGroup)
	for _, clerk := range clerks {
		wg.Add(1)
		clerk := clerk
		go func() {
			clerk.SetState(&pb.SetStateArgs{Epoch: epoch, State: reply.State})
		}()
	}
	wg.Wait()

	// Write to config service saying the new servers have up-to-date state.
	configCk.WriteConfig(epoch, servers)

	// Tell one of the servers to become primary.
	clerks[0].BecomePrimary(&pb.BecomePrimaryArgs{Epoch: epoch, Replicas: servers})
	return e.None
}
