package bank

import (
	"github.com/goose-lang/goose/machine"
	"github.com/mit-pdos/gokv/kv"
	"github.com/mit-pdos/gokv/lockservice"
	"github.com/tchajed/marshal"
)

// The maximum money supply, initially will all belong to accts[0]
const BAL_TOTAL = uint64(1000)

type BankClerk struct {
	lck   *lockservice.LockClerk
	kvck  kv.Kv
	accts []string
}

func acquire_two_good(lck *lockservice.LockClerk, l1, l2 string) {
	if l1 < l2 {
		lck.Lock(l1)
		lck.Lock(l2)
	} else {
		lck.Lock(l2)
		lck.Lock(l1)
	}
	return
}

func acquire_two(lck *lockservice.LockClerk, l1, l2 string) {
	// FIXME: need to add string lt to operational semantics
	lck.Lock(l1)
	lck.Lock(l2)
	return
}

func release_two(lck *lockservice.LockClerk, l1, l2 string) {
	lck.Unlock(l1)
	lck.Unlock(l2)
	return
}

func encodeInt(a uint64) string {
	return string(marshal.WriteInt(nil, a))
}

func decodeInt(a string) uint64 {
	v, _ := marshal.ReadInt([]byte(a))
	return v
}

// Requires that the account numbers are smaller than num_accounts
// If account balance in acc_from is at least amount, transfer amount to acc_to
func (bck *BankClerk) transfer_internal(acc_from string, acc_to string, amount uint64) {
	acquire_two(bck.lck, acc_from, acc_to)
	old_amount := decodeInt(bck.kvck.Get(acc_from))

	if old_amount >= amount {
		bck.kvck.Put(acc_from, string(encodeInt(old_amount-amount)))
		bck.kvck.Put(acc_to, encodeInt(decodeInt(bck.kvck.Get(acc_to))+amount))
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
		sum = sum + decodeInt(bck.kvck.Get(acct))
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

func MakeBankClerkSlice(lck *lockservice.LockClerk, kv kv.Kv, init_flag string, accts []string) *BankClerk {
	bck := new(BankClerk)
	bck.lck = lck
	bck.kvck = kv
	bck.accts = accts

	bck.lck.Lock(init_flag)
	// If init_flag has an empty value, initialize the accounts and set the flag.
	if bck.kvck.Get(init_flag) == "" {
		bck.kvck.Put(bck.accts[0], encodeInt(BAL_TOTAL))
		for _, acct := range bck.accts[1:] {
			bck.kvck.Put(acct, encodeInt(0))
		}
		bck.kvck.Put(init_flag, "1")
	}
	bck.lck.Unlock(init_flag)

	return bck
}

func MakeBankClerk(lck *lockservice.LockClerk, kv kv.Kv, init_flag string, acc1 string, acc2 string) *BankClerk {
	var accts []string
	accts = append(accts, acc1)
	accts = append(accts, acc2)
	return MakeBankClerkSlice(lck, kv, init_flag, accts)
}
