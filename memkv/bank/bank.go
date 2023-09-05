package bank

import (
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/memkv"
	"github.com/mit-pdos/gokv/memkv/lockservice"
	"github.com/tchajed/goose/machine"
)

// The maximum money supply, initially will all belong to accts[0]
const BAL_TOTAL = uint64(1000)

type BankClerk struct {
	lck   *lockservice.LockClerk
	kvck  *memkv.SeqKVClerk
	accts []uint64
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
	lck.Unlock(l1)
	lck.Unlock(l2)
	return
}

// Requires that the account numbers are smaller than num_accounts
// If account balance in acc_from is at least amount, transfer amount to acc_to
func (bck *BankClerk) transfer_internal(acc_from uint64, acc_to uint64, amount uint64) {
	acquire_two(bck.lck, acc_from, acc_to)
	old_amount := memkv.DecodeUint64(bck.kvck.Get(acc_from))

	if old_amount >= amount {
		bck.kvck.Put(acc_from, memkv.EncodeUint64(old_amount-amount))
		bck.kvck.Put(acc_to, memkv.EncodeUint64(memkv.DecodeUint64(bck.kvck.Get(acc_to))+amount))
	}
	release_two(bck.lck, acc_from, acc_to)
}

func (bck *BankClerk) SimpleTransfer() {
	for {
		src := machine.RandomUint64()
		dst := machine.RandomUint64()
		amount := machine.RandomUint64()
		if src < uint64(len(bck.accts)) && dst < uint64(len(bck.accts)) && src != dst {
			bck.transfer_internal(bck.accts[src], bck.accts[dst], amount)
		}
	}
}

func (bck *BankClerk) get_total() uint64 {
	var sum uint64

	// For deadlock avoidance, assume bck.accts is sorted
	for _, acct := range bck.accts {
		bck.lck.Lock(acct)
		sum = sum + memkv.DecodeUint64(bck.kvck.Get(acct))
	}

	for _, acct := range bck.accts {
		bck.lck.Unlock(acct)
	}

	return sum
}

func (bck *BankClerk) SimpleAudit() {
	for {
		if bck.get_total() != BAL_TOTAL {
			panic("Balance total invariant violated")
		}
	}
}

func MakeBankClerkSlice(lockhost memkv.HostName, kvhost memkv.HostName, cm *connman.ConnMan, init_flag uint64, accts []uint64, cid uint64) *BankClerk {
	bck := new(BankClerk)
	bck.lck = lockservice.MakeLockClerk(lockhost, cm)
	bck.kvck = memkv.MakeSeqKVClerk(kvhost, cm)
	bck.accts = accts

	bck.lck.Lock(init_flag)
	// If init_flag has an empty value, initialize the accounts and set the flag.
	if std.BytesEqual(bck.kvck.Get(init_flag), make([]byte, 0)) {
		bck.kvck.Put(bck.accts[0], memkv.EncodeUint64(BAL_TOTAL))
		for _, acct := range bck.accts[1:] {
			bck.kvck.Put(acct, memkv.EncodeUint64(0))
		}
		bck.kvck.Put(init_flag, make([]byte, 1))
	}
	bck.lck.Unlock(init_flag)

	return bck
}

func MakeBankClerk(lockhost memkv.HostName, kvhost memkv.HostName, cm *connman.ConnMan, init_flag uint64, acc1 uint64, acc2 uint64, cid uint64) *BankClerk {
	var accts []uint64
	accts = append(accts, acc1)
	accts = append(accts, acc2)
	return MakeBankClerkSlice(lockhost, kvhost, cm, init_flag, accts, cid)
}
