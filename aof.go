package gokv

import (
	"os"
	"syscall"
	// "fmt"
	"sync"
	// "time"
	// "math/rand"
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
		// numSyncs := 0
		// lastWritten := uint64(0)
		for {
			if len(a.membuf) == 0 {
				a.lengthCond.Wait()
				continue
			}

			// if uint64(len(a.membuf)) < 3 * lastWritten / 4 {
			// a.mu.Unlock()
			// time.Sleep(time.Duration((rand.Uint64() % 500)) * time.Microsecond)
			// a.mu.Lock()
			// }
			l := a.membuf
			a.membuf = make([]byte, 65536/2)
			a.mu.Unlock()

			a.f.Write(append(l))
			syscall.Fdatasync(int(a.f.Fd()))

			a.mu.Lock()
			a.durableLength += uint64(len(l))
			a.durableCond.Broadcast()
			// time.Sleep(time.Duration(rand.Uint64() % 500) * time.Microsecond)
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
	// time.Sleep(time.Duration((rand.Uint64() % 20)) * time.Microsecond)
	a.mu.Lock()
	for a.durableLength < length {
		a.durableCond.Wait()
	}
	a.mu.Unlock()
}
