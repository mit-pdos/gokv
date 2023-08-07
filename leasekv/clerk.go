package leasekv

// A purely client-side implementation of lease-based client caching for a
// key-value store.
import "github.com/mit-pdos/gokv/grove_ffi"

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
	cache map[string]cacheValue
}

func (k *LeaseKv) Get(key string) string {
	cv, ok := k.cache[key]
	if ok {
		low, _ := grove_ffi.GetTimeRange()
		if cv.l < low {
			return k.cache[key].v
		}
	}

	delete(k.cache, key)
	return k.kv.Get(key)
}

func DecodeValue(v string) (string, uint64) {
	panic("TODO")
}

func EncodeValue(v string, l uint64) string {
	panic("TODO")
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func (k *LeaseKv) GetAndCache(key string, cachetime uint64) string {
	for {
		enc := k.kv.Get(key)
		v, oldLeaseExpiration := DecodeValue(enc)
		_, latest := grove_ffi.GetTimeRange()
		var newLeaseExpiration = max(latest + cachetime, oldLeaseExpiration)

		// Try to update the lease expiration time
		resp := k.kv.ConditionalPut(key, enc, EncodeValue(v, newLeaseExpiration))
		if resp == "ok" {
			k.cache[key] = cacheValue{v: v, l: newLeaseExpiration}
			break
		}
	}
	return k.cache[key].v
}
