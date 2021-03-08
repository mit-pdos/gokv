package gokv

import (
	"github.com/tchajed/marshal"
	"sync"
)

type GoKVServer struct {
	mu        *sync.Mutex
	lastReply map[uint64]uint64
	lastSeq   map[uint64]uint64
	kvs       map[uint64]string
	durable   bool
	opLog     *AppendableFile
}

type PutArgs struct {
	Key   uint64
	Value string
}

func encodePut(args *PutArgs) []byte {
	v := []byte(args.Value)
	num_bytes := uint64(8 + 8 + len(v)) // key + value-len + value
	e := marshal.NewEnc(num_bytes)
	e.PutInt(args.Key)
	e.PutInt(uint64(len(v)))
	e.PutBytes(v)

	return e.Finish()
}

func (s *GoKVServer) PutRPC(args *PutArgs, reply *struct{}) error {
	s.mu.Lock()
	s.kvs[args.Key] = args.Value
	l := s.opLog.Append(encodePut(args))
	s.mu.Unlock()
	s.opLog.WaitAppend(l)
	return nil
}

func (s *GoKVServer) GetRPC(key *uint64, value *string) error {
	s.mu.Lock()
	*value = s.kvs[*key]
	s.mu.Unlock()
	return nil
}

// TODO: recovery

func MakeGoKVServer() *GoKVServer {
	srv := new(GoKVServer)
	srv.mu = new(sync.Mutex)
	srv.lastReply = make(map[uint64]uint64)
	srv.lastSeq = make(map[uint64]uint64)
	srv.kvs = make(map[uint64]string)
	srv.opLog = CreateAppendableFile("kvdur_log")
	return srv
}
