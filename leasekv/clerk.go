package leasekv

// A purely client-side implementation of lease-based client caching for a
// key-value store.
import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
)

type Kv struct {
	Put            func(key, value string)
	Get            func(key string) string
	ConditionalPut func(key, expect, value string) string
}

type cacheValue struct {
	v string
	l uint64
}

type LeaseKv struct {
	kv    *Kv
	mu    *sync.Mutex
	cache map[string]cacheValue
}

func DecodeValue(v string) cacheValue {
	panic("TODO")
}

func EncodeValue(c cacheValue) string {
	panic("TODO")
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func (k *LeaseKv) Get(key string) string {
	k.mu.Lock()
	cv, ok := k.cache[key]
	_, high := grove_ffi.GetTimeRange()
	if ok && high < cv.l {
		k.mu.Unlock()
		return cv.v
	}

	delete(k.cache, key)
	k.mu.Unlock()
	return DecodeValue(k.kv.Get(key)).v
}

func (k *LeaseKv) GetAndCache(key string, cachetime uint64) string {
	for {
		enc := k.kv.Get(key)
		old := DecodeValue(enc)

		_, latest := grove_ffi.GetTimeRange()
		var newLeaseExpiration = max(latest+cachetime, old.l)

		// Try to update the lease expiration time
		resp := k.kv.ConditionalPut(key, enc, EncodeValue(cacheValue{v: old.v, l: newLeaseExpiration}))
		if resp == "ok" {
			k.mu.Lock()
			k.cache[key] = cacheValue{v: old.v, l: newLeaseExpiration}
			break
		}
	}
	ret := k.cache[key].v
	k.mu.Unlock()
	return ret
}

func (k *LeaseKv) Put(key, val string) {
	for {
		enc := k.kv.Get(key)
		leaseExpiration := DecodeValue(enc).l

		earliest, _ := grove_ffi.GetTimeRange()
		if leaseExpiration > earliest {
			continue
		}
		// the lease has expired, so do the Put
		resp := k.kv.ConditionalPut(key, enc, EncodeValue(cacheValue{v: val, l: 0}))
		if resp == "ok" {
			break
		}
	}
}
