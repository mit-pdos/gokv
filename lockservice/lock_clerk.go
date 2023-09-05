package lockservice

import (
	"github.com/mit-pdos/gokv/kv"
)

type LockClerk struct {
	kv *kv.Kv
}

func (ck *LockClerk) Lock(key string) {
	for ck.kv.ConditionalPut(key, "", "1") != "ok" {
	}
}

func (ck *LockClerk) Unlock(key string) {
	ck.kv.Put(key, "")
}

func MakeLockClerk(kv *kv.Kv) *LockClerk {
	return &LockClerk{
		kv: kv,
	}
}
