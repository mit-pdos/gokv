package main

import "fmt"

type Deferred[S any, F any] struct {
	promise *Promise[S, F]

	resolve func(result *PromiseResult[S, F])

	reject func(err error)
}

func NewDeferred[S any, F any](eventLoop *EventLoop) *Deferred[S, F] {
	deferred := &Deferred[S, F]{}
	promise := NewPromise(func(resolveFunc func(result *PromiseResult[S, F]), rejectFunc func(err error)) {
		deferred.resolve = resolveFunc
		deferred.reject = rejectFunc
	}, eventLoop)
	deferred.promise = promise
	return deferred
}

type Board struct {
	cards          []string
	mapChangeWatch *Deferred[uint64, uint64]
}

func notifyMap(b *Board, eventLoop *EventLoop) {
	b.mapChangeWatch.resolve(nil)
	Debug("creating new deferred")
	b.mapChangeWatch = NewDeferred[uint64, uint64](eventLoop)
}

func mapBoard(b *Board, f func(card string) *Promise[string, string], eventLoop *EventLoop) *Promise[string, string] {
	cardPromises := make(map[string]string)
	completedPromises := 0
	totalPromises := 0
	newTextsComputed := NewDeferred[string, string](eventLoop)
	for _, card := range b.cards {
		// see if we already made promise for this card
		_, ok := cardPromises[card]
		if ok {
			continue
		}

		totalPromises++
		cardPromises[card] = "placeholder"

		// make new promise for this card
		then(f(card), func(result *PromiseResult[string, string]) (*PromiseResult[string, string], error) {
			// annoying case because can't use generics
			cardPromises[card] = result.successValue
			completedPromises++
			if completedPromises == totalPromises {
				newTextsComputed.resolve(nil)
			}
			return nil, nil
		}, nil, eventLoop)
	}

	boardChanged := false
	then(b.mapChangeWatch.promise, func(result *PromiseResult[uint64, uint64]) (*PromiseResult[uint64, uint64], error) {
		boardChanged = true
		return nil, nil
	}, nil, eventLoop)
	return then(newTextsComputed.promise, func(result *PromiseResult[string, string]) (*PromiseResult[string, string], error) {
		// now we can safely mutate the board
		if boardChanged {
			return &PromiseResult[string, string]{nestedPromise: mapBoard(b, f, eventLoop)}, nil
		}

		changedBoard := false
		for i, card := range b.cards {
			newText := cardPromises[card]
			changedBoard = changedBoard || card != newText
			b.cards[i] = newText
		}
		if changedBoard {
			notifyMap(b, eventLoop)
		}

		return nil, nil
	}, nil, eventLoop)
}

func MapBoardExample(eventLoop *EventLoop) {
	board := &Board{
		cards:          []string{"red", "green", "blue"},
		mapChangeWatch: NewDeferred[uint64, uint64](eventLoop),
	}

	fmt.Println("Starting board", board.cards)

	firstMap := mapBoard(board, func(card string) *Promise[string, string] {
		deferred := NewDeferred[string, string](eventLoop)
		deferred.resolve(&PromiseResult[string, string]{successValue: "yellow"})
		return deferred.promise
	}, eventLoop)

	then(firstMap, func(result *PromiseResult[string, string]) (*PromiseResult[uint64, uint64], error) {
		fmt.Println("Ending board", board.cards)
		return nil, nil
	}, nil, eventLoop)
}
