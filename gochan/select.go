package gochan

// This file defines helper functions to interpret select statements. It is for
// convenience in chan_test.go, it is not intended to be used in any Goose
// translation.

import (
	"math/rand"
)

type SelectCase interface {
	trySelect() bool
}

// FIXME: TestSelfSelect doesn't terminate currently. The problem is that
// concurrent NonblockingReceive and NonblockingSend will never match up.
//
// A fix: a blocking select statement shouldn't repeatedly run
// "Nonblocking{Send,Receive}". Instead, for liveness, it could speculatively
// run the blocking version of {Send,Receive}. Basically want to use
// angelic choice to pick a select case.

func Select(blocking bool, cases ...SelectCase) int {
	i := rand.Int() % len(cases)
	if cases[i].trySelect() {
		return i
	}

	for {
		for i := 0; i < len(cases); i++ {
			if cases[i].trySelect() {
				return i
			}
		}
		if !blocking {
			return -1
		}
	}
}

type CaseSend[T any] struct {
	ch *Channel[T]
	v T
}

func NewCaseSend[T any](ch *Channel[T], v T) *CaseSend[T] {
	return &CaseSend[T]{
		ch: ch,
		v: v,
	}
}

type CaseReceive[T any] struct {
	ch *Channel[T]
	Val T
	Ok bool
}

func NewCaseReceive[T any](ch *Channel[T]) *CaseReceive[T] {
	return &CaseReceive[T]{
		ch: ch,
	}
}

func (c *CaseReceive[T]) trySelect() (selected bool) {
	selected, c.Val, c.Ok = c.ch.NonblockingReceive()
	return
}

func (c *CaseSend[T]) trySelect() bool {
	return c.ch.NonblockingSend(c.v)
}
