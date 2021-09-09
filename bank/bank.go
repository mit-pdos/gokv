package bank

import (
	"github.com/mit-pdos/gokv/lockservice"
	"github.com/mit-pdos/gokv/memkv"
)

type BankClerk struct {
	lck  *lockservice.LockClerk
	kvck *memkv.MemKVClerk
	acc1 uint64
	acc2 uint64
}

func acquire_two(lck *lockservice.LockClerk, l1 uint64, l2 uint64) {
	if l1 < l2 {
		lck.Lock(l1)
		lck.Lock(l2)
	} else {
		lck.Lock(l2)
		lck.Lock(l1)
	}
	return
}

func release_two(lck *lockservice.LockClerk, l1 uint64, l2 uint64) {
	if l1 < l2 {
		lck.Unlock(l2)
		lck.Unlock(l1)
	} else {
		lck.Unlock(l1)
		lck.Unlock(l2)
	}
	return
}

// Requires that the account numbers are smaller than num_accounts
// If account balance in acc_from is at least amount, transfer amount to acc_to
func (bck *BankClerk) transfer_internal(acc_from uint64, acc_to uint64, amount uint64) {
	acquire_two(bck.lck, acc_from, acc_to)
	old_amount := memkv.DecodeUint64(bck.kvck.Get(acc_from))

	if old_amount >= amount {
		bck.kvck.Put(acc_from, memkv.EncodeUint64(old_amount-amount))
		bck.kvck.Put(acc_to, memkv.EncodeUint64(memkv.DecodeUint64(bck.kvck.Get(acc_to)) + amount))
	}
	release_two(bck.lck, acc_from, acc_to)
}

func (bck *BankClerk) SimpleTransfer(amount uint64) {
	bck.transfer_internal(bck.acc1, bck.acc2, amount)
}

// If account balance in acc_from is at least amount, transfer amount to acc_to
func (bck *BankClerk) SimpleAudit() uint64 {
	acquire_two(bck.lck, bck.acc1, bck.acc2)
	sum := memkv.DecodeUint64(bck.kvck.Get(bck.acc1)) + memkv.DecodeUint64(bck.kvck.Get(bck.acc2))
	release_two(bck.lck, bck.acc1, bck.acc2)
	return sum
}

func MakeBankClerk(lockhost memkv.HostName, kvhost memkv.HostName, acc1 uint64, acc2 uint64, cid uint64) *BankClerk {
	bck := new(BankClerk)
	bck.lck = lockservice.MakeLockClerk(lockhost)
	bck.kvck = memkv.MakeMemKVClerk(kvhost)
	bck.acc1 = acc1
	bck.acc2 = acc2
	return bck
}
