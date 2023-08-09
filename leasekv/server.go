package leasekv

// A KV server with support for clients caches with leases.

import "github.com/mit-pdos/gokv/grove_ffi"

type LeaseKvServer struct {
	kvs       map[string]string
	leases    map[string]uint64
	stagedKvs map[string]string
}

// Rule: reads must never wait. Puts can wait if there's a lease on the key
// being written.

func (s *LeaseKvServer) resolveStagedValue(key string) {
	stagedVal, isStaged := s.stagedKvs[key]
	if isStaged {
		earliest, latest := grove_ffi.GetTimeRange()
		// FIXME: deal with the case that key not in s.leases (i.e. no active lease)
		if s.leases[key] >= latest {
			// if lease is expired, then set the current value to the stagedVal
			s.kvs[key] = stagedVal
			delete(s.leases, key)
			delete(s.stagedKvs, key)
		} else if s.leases[key] <= earliest {
			// if lease is still valid, then the staged value should not take
			// effect
		} else {
			// if lease is not valid, but the new value has not taken over,
			// return the new value and tell the client to wait
			panic("TODO: wait")
		}
	}
}

func (s *LeaseKvServer) Get(key string) string {
	s.resolveStagedValue(key)
	return s.kvs[key]
}

func (s *LeaseKvServer) GetAndCache(key string, cachetime uint64) (string, uint64) {
	_, latest := grove_ffi.GetTimeRange()
	newLeaseExpiration := latest + cachetime
	leaseExpiration, ok := s.leases[key]
	if ok && newLeaseExpiration < leaseExpiration {
		return s.kvs[key], leaseExpiration
	} else {
		s.leases[key] = newLeaseExpiration
		return s.kvs[key], newLeaseExpiration
	}
}

func (s *LeaseKvServer) Put(key string, val string) uint64 {
	panic("TODO")
}
