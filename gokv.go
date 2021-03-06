package gokv

import (
	"github.com/tchajed/marshal"
	"io/ioutil"
	"sync"
)

type GoKVServer struct {
	mu        *sync.Mutex
	lastReply map[uint64]uint64
	lastSeq   map[uint64]uint64
	kvs       map[uint64]string
	kvsSize   uint64
}

type PutArgs struct {
	Key   uint64
	Value string
}

func (s *GoKVServer) PutRPC(args PutArgs, reply *struct{}) error {
	s.mu.Lock()
	oldv, ok := s.kvs[args.Key]
	if ok {
		s.kvsSize -= uint64(len([]byte(oldv)))
		s.kvsSize -= 8 // "value size" size
		s.kvsSize -= 8 // Key size
	}
	s.kvs[args.Key] = args.Value
	s.kvsSize += uint64(len(s.kvs[args.Key]))
	s.kvsSize += 8
	s.kvsSize += 8
	WriteDurableKVServer(s)
	s.mu.Unlock()
	return nil
}

func (s *GoKVServer) GetRPC(key *uint64, value *string) error {
	s.mu.Lock()
	*value = s.kvs[*key]
	s.mu.Unlock()
	return nil
}

func EncMap(e *marshal.Enc, m map[uint64]uint64) {
	e.PutInt(uint64(len(m)))
	for key, value := range m {
		e.PutInt(key)
		e.PutInt(value)
	}
}

// requires ?some? amount of space
func EncByteMap(e *marshal.Enc, m map[uint64]string) {
	e.PutInt(uint64(len(m)))
	for key, value := range m {
		e.PutInt(key)
		e.PutInt(uint64(len(value)))
		e.PutBytes([]byte(value))
	}
}

func WriteDurableKVServer(ks *GoKVServer) {
	// TODO: need to account for size of values; probably best to enforce a limit on value size
	num_bytes := uint64(8*(2*len(ks.lastSeq)+2*len(ks.lastReply) + 2)) + ks.kvsSize
	e := marshal.NewEnc(num_bytes) // 4 uint64s
	EncMap(&e, ks.lastSeq)
	EncMap(&e, ks.lastReply)
	EncByteMap(&e, ks.kvs)

	// TODO: this isn't atomic
	ioutil.WriteFile("kvdur", e.Finish(), 0644)
	return
}

func MakeGoKVServer() *GoKVServer {
	srv := new(GoKVServer)
	srv.mu = new(sync.Mutex)
	srv.lastReply = make(map[uint64]uint64)
	srv.lastSeq = make(map[uint64]uint64)
	srv.kvs = make(map[uint64]string)
	srv.kvsSize = 8 // for len
	return srv
}
