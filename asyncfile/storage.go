package asyncfile

import (
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
)

type AsyncFile struct {
	mu               *sync.Mutex
	data             []byte
	filename         string
	index            uint64
	indexCond        *sync.Cond
	durableIndex     uint64
	durableIndexCond *sync.Cond

	closeRequested bool
	closed         bool
	closedCond     *sync.Cond
}

func (s *AsyncFile) Write(data []byte) func() {
	// XXX: can read index here because it's owned by the owner of the File object
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = data
	s.index = std.SumAssumeNoOverflow(s.index, 1)
	index := s.index
	s.indexCond.Signal()
	return func() { s.wait(index) }
}

func (s *AsyncFile) wait(index uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for s.durableIndex < index {
		s.durableIndexCond.Wait()
	}
}

func (s *AsyncFile) flushThread() {
	s.mu.Lock()
	for {
		if s.closeRequested {
			// flush everything and exit
			grove_ffi.FileWrite(s.filename, s.data)
			s.durableIndex = s.index
			s.durableIndexCond.Broadcast()
			s.closed = true
			s.mu.Unlock()
			s.closedCond.Signal()
			return
		}
		if s.durableIndex >= s.index {
			s.indexCond.Wait()
			continue
		}
		index := s.index
		data := s.data
		s.mu.Unlock()
		grove_ffi.FileWrite(s.filename, data)
		s.mu.Lock()
		s.durableIndex = index
		s.durableIndexCond.Broadcast()
		// TODO: can avoid false wakeups by having two condvars like aof
	}
}

func (s *AsyncFile) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closeRequested = true
	s.indexCond.Signal()
	for !s.closed {
		s.closedCond.Wait()
	}
}

// returns the state, then the File object
func MakeAsyncFile(filename string) ([]byte, *AsyncFile) {
	var mu sync.Mutex
	s := &AsyncFile{
		mu:               &mu,
		indexCond:        sync.NewCond(&mu),
		closedCond:       sync.NewCond(&mu),
		durableIndexCond: sync.NewCond(&mu),
		filename:         filename,
		data:             grove_ffi.FileRead(filename),
		index:            0,
		durableIndex:     0,
		closed:           false,
		closeRequested:   false,
	}
	data := s.data
	go s.flushThread()
	return data, s
}
