package main

import "fmt"

type Deferred struct {
	promise *Promise

	resolve func(interface{})

	reject func(interface{})
}

func (r *Runtime) NewDeferred() *Deferred {
	deferred := &Deferred{}
	promise := r.NewPromise(func(resolveFunc func(interface{}), rejectFunc func(interface{})) {
		deferred.resolve = resolveFunc
		deferred.reject = rejectFunc
	})
	deferred.promise = promise
	return deferred
}

type Board struct {
	cards          []string
	mapChangeWatch *Deferred
}

func (r *Runtime) notifyMap(b *Board) {
	b.mapChangeWatch.resolve(nil)
	b.mapChangeWatch = r.NewDeferred()
}

func (r *Runtime) mapBoard(b *Board, f func(card string) *Promise) *Promise {
	cardPromises := make(map[string]string)
	completedPromises := 0
	totalPromises := 0
	newTextsComputed := r.NewDeferred()
	for _, card := range b.cards {
		// see if we already made promise for this card
		_, ok := cardPromises[card]
		if ok {
			continue
		}

		totalPromises++
		cardPromises[card] = "placeholder"

		// make new promise for this card
		r.then(f(card), func(result interface{}) interface{} {
			// annoying case because can't use generics
			stringResult, ok := result.(string)
			if !ok {
				stringResult = "need to add err handling in promise"
			}
			cardPromises[card] = stringResult
			completedPromises++
			if completedPromises == totalPromises {
				newTextsComputed.resolve(nil)
			}
			return nil
		}, nil)
	}

	boardChanged := false
	r.then(b.mapChangeWatch.promise, func(result interface{}) interface{} {
		boardChanged = true
		return nil
	}, nil)
	return r.then(newTextsComputed.promise, func(result interface{}) interface{} {
		// now we can safely mutate the board
		if boardChanged {
			return r.mapBoard(b, f)
		}

		changedBoard := false
		for i, card := range b.cards {
			newText := cardPromises[card]
			changedBoard = changedBoard || card != newText
			b.cards[i] = newText
		}
		if changedBoard {
			r.notifyMap(b)
		}

		return nil
	}, nil)
}

func (r *Runtime) MapBoardExample() {
	board := &Board{
		cards:          []string{"red", "green", "blue"},
		mapChangeWatch: r.NewDeferred(),
	}

	fmt.Println("Starting board", board.cards)

	firstMap := r.mapBoard(board, func(card string) *Promise {
		promise := r.NewPromise(func(resolveFunc func(interface{}), rejectFunc func(interface{})) {})
		r.resolve(promise, "yellow")
		return promise
	})

	r.then(firstMap, func(result interface{}) interface{} {
		fmt.Println("Ending board", board.cards)
		return nil
	}, nil)
}
