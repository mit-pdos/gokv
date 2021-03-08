package gokv

import (
	"os"
	"syscall"
	"sync"
)

type AppendableFile struct {
	f *os.File
	mu *sync.Mutex

	durableCond *sync.Cond
	lengthCond *sync.Cond

	membuf []byte
	durableLength uint64
}

func CreateAppendableFile(fname string) *AppendableFile {
	a := new(AppendableFile)
	a.mu = new(sync.Mutex)
	a.lengthCond = sync.NewCond(a.mu)
	a.durableCond = sync.NewCond(a.mu)

	var err error
	a.f, err = os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	go func() {
		a.mu.Lock()
		for {
			if len(a.membuf) == 0 {
				a.lengthCond.Wait()
				continue
			}

			l := a.membuf
			a.membuf = make([]byte, 0)
			a.mu.Unlock()

			// TODO: replace this with a AtomicAppend() call
			a.f.Write(append(l))
			syscall.Fdatasync(int(a.f.Fd()))

			a.mu.Lock()
			a.durableLength += uint64(len(l))
			a.durableCond.Broadcast()
		}
	}()

	return a
}

func (a *AppendableFile) Append(data []byte) uint64 {
	a.mu.Lock()
	a.membuf = append(a.membuf, data...)
	r := a.durableLength + uint64(len(a.membuf))
	a.lengthCond.Signal()
	a.mu.Unlock()
	return r
}

func (a *AppendableFile) WaitAppend(length uint64) {
	a.mu.Lock()
	for a.durableLength < length {
		a.durableCond.Wait()
	}
	a.mu.Unlock()
}
