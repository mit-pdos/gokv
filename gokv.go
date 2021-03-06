package gokv

import (
	"sync"
	"github.com/tchajed/marshal"
	"io/ioutil"
)

type GoKVServer struct {
	mu *sync.Mutex
	lastReply map[uint64]uint64
	lastSeq map[uint64]uint64
	kvs map[uint64]string
}

type PutArgs struct {
	Key uint64
	Value string
}

func (s *GoKVServer) PutRPC(args PutArgs, reply *struct{}) error {
	s.mu.Lock()
	s.kvs[args.Key] = args.Value
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
	num_bytes := uint64(8 * (2*len(ks.lastSeq) + 2*len(ks.lastReply)) + (8 + 32)*len(ks.kvs) + 3*8)
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
	return srv
}
