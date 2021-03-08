package gokv

import (
	"os"
	"syscall"
	"time"
	"sync"
)

type AppendableFile struct {
	f *os.File
	mu *sync.Mutex

	durableCond *sync.Cond
	lengthCond *sync.Cond
	length uint64
	durableLength uint64
}

func CreateAppendableFile(fname string) *AppendableFile {
	a := new(AppendableFile)
	a.mu = new(sync.Mutex)
	a.lengthCond = sync.NewCond(a.mu)
	a.durableCond = sync.NewCond(a.mu)

	var err error
	a.f, err = os.OpenFile(fname, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	go func() {
		a.mu.Lock()
		for {
			if a.length == a.durableLength {
				a.lengthCond.Wait()
				continue
			}
			a.mu.Unlock()
			time.Sleep(2000 * time.Microsecond)

			a.mu.Lock()
			l := a.length // TODO: do atomic load
			a.mu.Unlock()

			syscall.Fdatasync(int(a.f.Fd()))

			a.mu.Lock()
			a.durableLength = l
			a.durableCond.Broadcast()
		}
	}()

	return a
}

// Not safe to do concurrent Append
func (a *AppendableFile) Append(data []byte) {
	a.mu.Lock()
	_, err := a.f.Write(data)
	if err != nil {
		panic(err)
	}
	a.length += uint64(len(data))
	a.lengthCond.Signal()

	for a.length > a.durableLength {
		a.durableCond.Wait()
	}
	a.mu.Unlock()
}
