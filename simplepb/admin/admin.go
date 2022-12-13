package admin

import (
	"log"
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"github.com/mit-pdos/gokv/simplepb/e"
	"github.com/mit-pdos/gokv/simplepb/pb"
	"github.com/tchajed/goose/machine"
)

func EnterNewConfig(configHost grove_ffi.Address, servers []grove_ffi.Address) e.Error {
	if len(servers) == 0 {
		log.Println("Tried creating empty config")
		return e.EmptyConfig
	}

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
		log.Printf("Error while getting state and sealing in epoch %d", epoch)
		return reply.Err
	}

	// Set the state of all the new servers.
	clerks := make([]*pb.Clerk, len(servers))
	var i = uint64(0)
	for i < uint64(len(clerks)) {
		clerks[i] = pb.MakeClerk(servers[i])
		i += 1
	}
	wg := new(sync.WaitGroup)
	errs := make([]e.Error, len(clerks))

	i = 0
	for i < uint64(len(clerks)) {
		wg.Add(1)
		clerk := clerks[i]
		locali := i
		go func() {
			errs[locali] = clerk.SetState(&pb.SetStateArgs{Epoch: epoch, State: reply.State, NextIndex: reply.NextIndex})
			wg.Done()
		}()
		i += 1
	}
	wg.Wait()

	var err = e.None
	i = 0
	for i < uint64(len(errs)) {
		err2 := errs[i]
		if err2 != e.None {
			err = err2
		}
		i += 1
	}
	if err != e.None {
		log.Println("Error while setting state and entering new epoch")
		return err
	}

	// Write to config service saying the new servers have up-to-date state.
	if configCk.WriteConfig(epoch, servers) != e.None {
		log.Println("Error while writing to config service")
		return e.Stale
	}

	// Tell one of the servers to become primary.
	clerks[0].BecomePrimary(&pb.BecomePrimaryArgs{Epoch: epoch, Replicas: servers})
	return e.None
}
