package cachekv

// A purely client-side implementation of lease-based client caching for a
// key-value store.
import (
	"sync"

	"github.com/mit-pdos/gokv/cachekv/cachevalue_gk"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/kv"
)

type CacheKv struct {
	kv    kv.KvCput
	mu    *sync.Mutex
	cache map[string]cachevalue_gk.S
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func Make(kv kv.KvCput) *CacheKv {
	return &CacheKv{
		kv:    kv,
		mu:    new(sync.Mutex),
		cache: make(map[string]cachevalue_gk.S),
	}
}

func (k *CacheKv) Get(key string) string {
	k.mu.Lock()
	cv, ok := k.cache[key]
	_, high := grove_ffi.GetTimeRange()
	if ok && high < cv.L {
		k.mu.Unlock()
		return cv.V
	}

	delete(k.cache, key)
	k.mu.Unlock()
	ret, _ := cachevalue_gk.Unmarshal([]byte(k.kv.Get(key)))
	return ret.V
}

func (k *CacheKv) GetAndCache(key string, cachetime uint64) string {
	for {
		enc := k.kv.Get(key)
		old, _ := cachevalue_gk.Unmarshal([]byte(enc))

		_, latest := grove_ffi.GetTimeRange()
		newLeaseExpiration := max(latest+cachetime, old.L)

		// Try to update the lease expiration time
		resp := k.kv.ConditionalPut(key, enc, string(cachevalue_gk.Marshal(make([]byte, 0), cachevalue_gk.S{V: old.V, L: newLeaseExpiration})))
		if resp == "ok" {
			k.mu.Lock()
			k.cache[key] = cachevalue_gk.S{V: old.V, L: newLeaseExpiration}
			break
		}
	}
	ret := k.cache[key].V
	k.mu.Unlock()
	return ret
}

func (k *CacheKv) Put(key, val string) {
	for {
		enc := k.kv.Get(key)
		cval, _ := cachevalue_gk.Unmarshal([]byte(enc))
		leaseExpiration := cval.L

		earliest, _ := grove_ffi.GetTimeRange()
		if leaseExpiration > earliest {
			continue
		}
		// the lease has expired, so do the Put
		resp := k.kv.ConditionalPut(key, enc, string(cachevalue_gk.Marshal(make([]byte, 0), cachevalue_gk.S{V: val, L: 0})))
		if resp == "ok" {
			break
		}
	}
}
