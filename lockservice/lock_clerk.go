package lockservice

import (
	"github.com/mit-pdos/gokv/memkv"
)

type LockClerk struct {
	kv *memkv.MemKVClerk
}

func (ck *LockClerk) Lock(key uint64) {
	for ck.kv.ConditionalPut(key, make([]byte, 0), make([]byte, 1)) {
	}
}

func (ck *LockClerk) Unlock(key uint64) {
	ck.kv.Put(key, make([]byte, 0))
}

func MakeLockClerk(lockhost memkv.HostName) *LockClerk {
	return &LockClerk{
		kv: memkv.MakeMemKVClerk(lockhost),
	}
}
