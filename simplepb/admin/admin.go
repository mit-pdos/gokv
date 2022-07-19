package admin

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/goose/machine"
	"sync"
)

func InitializeSystem(configHost grove_ffi.Address, servers []grove_ffi.Address) e.Error {
	configCk := config.MakeClerk(configHost)
	// Get new epoch number from config service.
	epoch, _ := configCk.GetEpochAndConfig()

	// Write to config service saying the new servers have up-to-date state.
	configCk.WriteConfig(epoch, servers)

	// Tell one of the servers to become primary.
	clerk := pb.MakeClerk(servers[0])
	clerk.BecomePrimary(&pb.BecomePrimaryArgs{Epoch: epoch, Replicas: servers})
	return e.None
}

func EnterNewConfig(configHost grove_ffi.Address, servers []grove_ffi.Address) e.Error {
	configCk := config.MakeClerk(configHost)
	// Get new epoch number from config service.
	// Read from config service, fenced with that epoch.
	epoch, oldServers := configCk.GetEpochAndConfig()

	// Enter new epoch on one of the old servers.
	// Get a copy of the state from that old server.

	// FIXME: this +1 is a terrible hack; liveness bug.
	// Should try all of the servers, starting from some random offset.
	id := (machine.RandomUint64() + 1) % uint64(len(oldServers))
	oldClerk := pb.MakeClerk(oldServers[id])
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
	errs := make([]e.Error, len(clerks))
	for i, clerk := range clerks {
		wg.Add(1)
		clerk := clerk
		i := i
		go func() {
			errs[i] = clerk.SetState(&pb.SetStateArgs{Epoch: epoch, State: reply.State})
			wg.Done()
		}()
	}
	wg.Wait()

	var err = e.None
	for _, err2 := range errs {
		if err2 != e.None {
			err = err2
		}
	}
	if err != e.None {
		return err
	}

	// Write to config service saying the new servers have up-to-date state.
	if configCk.WriteConfig(epoch, servers) != e.None {
		return e.Stale
	}

	// Tell one of the servers to become primary.
	clerks[0].BecomePrimary(&pb.BecomePrimaryArgs{Epoch: epoch, Replicas: servers})
	return e.None
}
