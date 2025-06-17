package gochan

import (
	"runtime"
	"sync"
	"testing"
)

type response struct {
}

type myError struct {
}

func (myError) Error() string { return "" }

func doRequest(useSelect bool) (*response, error) {
	type async struct {
		resp *response
		err  error
	}
	ch := NewChan[*async](0)
	done := NewChan[struct{}](0)

	if useSelect {
		go func() {
			switch Select(false,
				NewCaseSend(ch, &async{resp: nil, err: myError{}}),
				NewCaseReceive(done),
			) {
			case 0:
			case 1:
			}
		}()
	} else {
		go func() {
			ch.Send(&async{resp: nil, err: myError{}})
		}()
	}

	r, _ := ch.Receive()
	runtime.Gosched()
	return r.resp, r.err
}

func TestChanSendSelectBarrier(t *testing.T) {
	t.Parallel()
	testChanSendBarrier(true)
}

func TestChanSendBarrier(t *testing.T) {
	t.Parallel()
	testChanSendBarrier(false)
}

func testChanSendBarrier(useSelect bool) {
	var wg sync.WaitGroup
	outer := 2
	inner := 20
	for i := 0; i < outer; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var garbage []byte
			for j := 0; j < inner; j++ {
				_, err := doRequest(useSelect)
				_, ok := err.(myError)
				if !ok {
					panic(1)
				}
				garbage = makeByte()
			}
			_ = garbage
		}()
	}
	wg.Wait()
}

//go:noinline
func makeByte() []byte {
	return make([]byte, 1<<10)
}
