package goosekv

import (
	"github.com/mit-pdos/gokv/aof"
	"github.com/mit-pdos/lockservice/grove_common"
	"github.com/mit-pdos/lockservice/grove_ffi"
	"sync"
)

type GoKVShardServer struct {
	mu        *sync.Mutex
	lastReply map[uint64]uint64
	lastSeq   map[uint64]uint64

	// true iff we own that shard
	shardMap map[uint64]bool

	// if anything is in shardMap, then we have a map[] initialized in kvss
	kvss     map[uint64](map[uint64][]byte)
	opLog    *aof.AppendOnlyFile
}

type PutArgs struct {
	Key   uint64
	Value ValueType
}

func (s *GoKVShardServer) put_inner(args *PutRequest, reply *PutReply) {
	last, ok := s.lastSeq[args.CID]
	if ok && args.Seq <= last {
		reply.Err = s.lastReply[args.CID]
		return
	}
	s.lastSeq[args.CID] = args.Seq

	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		s.kvss[sid][args.Key] = args.Value // give ownership of the slice to the server
		s.lastReply[args.CID] = 0
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
	}
}

func (s *GoKVShardServer) Put(args *PutRequest, reply *PutReply) {
	s.mu.Lock()
	s.put_inner(args, reply)
	l := s.opLog.Append(encodePutRequest(args))
	s.mu.Unlock()
	s.opLog.WaitAppend(l)
}

// TODO: Get()
// TODO: InstallShard()
// TODO: recovery

func MakeGoKVShardServer() *GoKVShardServer {
	srv := new(GoKVShardServer)
	srv.mu = new(sync.Mutex)
	srv.lastReply = make(map[uint64]uint64)
	srv.lastSeq = make(map[uint64]uint64)
	srv.kvss = make(map[uint64]map[uint64][]byte)
	srv.shardMap = make(map[uint64]bool)
	srv.opLog = aof.CreateAppendOnlyFile("kvdur_log")
	return srv
}

func (gkv *GoKVShardServer) Start() {
	handlers := make(map[uint64]grove_common.RawRpcFunc)
	handlers[KV_PUT] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(PutReply)
		gkv.Put(decodePutRequest(rawReq), rep)
		*rawReply = encodePutReply(rep)
	}
	grove_ffi.StartRPCServer(handlers)
}
