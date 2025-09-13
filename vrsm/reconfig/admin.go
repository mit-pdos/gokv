package reconfig

import (
	"log"
	"sync"

	"github.com/goose-lang/primitive"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/reconfig/replica/becomeprimaryargs_gk"
	"github.com/mit-pdos/gokv/reconfig/replica/configuration_gk"
	"github.com/mit-pdos/gokv/vrsm/configservice"
	"github.com/mit-pdos/gokv/vrsm/configservice/config_gk"
	"github.com/mit-pdos/gokv/vrsm/replica"
	"github.com/mit-pdos/gokv/vrsm/replica/err_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/getstateargs_gk"
	"github.com/mit-pdos/gokv/vrsm/replica/setstateargs_gk"
)

func EnterNewConfig(configHosts []grove_ffi.Address, servers []grove_ffi.Address) err_gk.E {
	if len(servers) == 0 {
		log.Println("Tried creating empty config")
		return err_gk.EmptyConfig
	}

	configCk := configservice.MakeClerk(configHosts)
	// Get new epoch number from config service.
	// Read from config service, fenced with that epoch.
	epoch, conf := configCk.ReserveEpochAndGetConfig()
	oldServers := conf.Addrs
	log.Printf("Reserved %d", epoch)

	// Enter new epoch on one of the old servers.
	// Get a copy of the state from that old server.

	// TODO: maybe should try all of the servers, starting from some random
	// offset. This "+1" can also go away.
	id := (primitive.RandomUint64() + 1) % uint64(len(oldServers))
	oldClerk := replica.MakeClerk(oldServers[id])
	reply := oldClerk.GetState(&getstateargs_gk.S{Epoch: epoch})
	if reply.Err != err_gk.None {
		log.Printf("Error while getting state and sealing in epoch %d", epoch)
		return reply.Err
	}

	// FIXME: maybe use "makeClerks" helper function from simplepb/clerk
	// Set the state of all the new servers.
	clerks := make([]*replica.Clerk, len(servers))
	var i = uint64(0)
	for i < uint64(len(clerks)) {
		clerks[i] = replica.MakeClerk(servers[i])
		i += 1
	}
	wg := new(sync.WaitGroup)
	errs := make([]err_gk.E, len(clerks))

	i = 0
	for i < uint64(len(clerks)) {
		wg.Add(1)
		clerk := clerks[i]
		locali := i
		go func() {
			errs[locali] = clerk.SetState(&setstateargs_gk.S{Epoch: epoch, State: reply.State, NextIndex: reply.NextIndex, CommittedNextIndex: reply.CommittedNextIndex})
			wg.Done()
		}()
		i += 1
	}
	wg.Wait()

	var err = err_gk.None
	i = 0
	for i < uint64(len(errs)) {
		err2 := errs[i]
		if err2 != err_gk.None {
			err = err2
		}
		i += 1
	}
	if err != err_gk.None {
		log.Println("Error while setting state and entering new epoch")
		return err
	}

	// Write to config service saying the new servers have up-to-date state.
	if configCk.TryWriteConfig(epoch, config_gk.S{Addrs: servers}) != err_gk.None {
		log.Println("Error while writing to config service")
		return err_gk.Stale
	}

	// Tell one of the servers to become primary.
	clerks[0].BecomePrimary(&becomeprimaryargs_gk.S{Epoch: epoch, Conf: configuration_gk.S{Replicas: servers}})
	return err_gk.None
}
