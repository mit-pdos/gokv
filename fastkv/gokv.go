package fastkv

import (
	"github.com/mit-pdos/gokv/aof"
	"github.com/upamanyus/urpc/rpc"
	"sync"
)

type VersionedValue struct {
	val []byte
	ver uint64
}

type GoKVShardServer struct {
	mu        *sync.Mutex
	lastReply map[uint64]GetReply
	lastSeq   map[uint64]uint64

	// true iff we own that shard
	shardMap map[uint64]bool

	// if anything is in shardMap, then we have a map[] initialized in kvss
	kvss  map[uint64](map[uint64]VersionedValue)
	opLog *aof.AppendOnlyFile
}

type PutArgs struct {
	Key   uint64
	Value ValueType
}

func (s *GoKVShardServer) put_inner(args *PutRequest, reply *PutReply, l uint64) {
	last, ok := s.lastSeq[args.CID]
	if ok && args.Seq <= last {
		// XXX: this is a bit hacky
		reply.Err = s.lastReply[args.CID].Err
		return
	}
	s.lastSeq[args.CID] = args.Seq

	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		s.kvss[sid][args.Key] = VersionedValue{val:args.Value, ver:l} // give ownership of the slice to the server
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
	}

	// XXX: this is a bit hacky (same as above)
	s.lastReply[args.CID] = GetReply{Err:reply.Err}
}

func (s *GoKVShardServer) PutRPC(args *PutRequest, reply *PutReply) {
	s.mu.Lock()
	l := s.opLog.Append(encodePutRequest(args))
	s.put_inner(args, reply, l)
	s.mu.Unlock()
	s.opLog.WaitAppend(l)
}

func (s *GoKVShardServer) get_inner(args *GetRequest, reply *GetReply) {
	last, ok := s.lastSeq[args.CID]
	if ok && args.Seq <= last {
		*reply = s.lastReply[args.CID]
		return
	}
	s.lastSeq[args.CID] = args.Seq

	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		reply.Value = s.kvss[sid][args.Key].val
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
	}
	s.lastReply[args.CID] = *reply
}

func (s *GoKVShardServer) GetRPC(args *GetRequest, reply *GetReply) {
	s.mu.Lock()
	s.get_inner(args, reply)
	l := s.opLog.Append(encodeGetRequest(args))
	s.mu.Unlock()
	s.opLog.WaitAppend(l)
}

func (s *GoKVShardServer) FastUnsafeGetRPC(args *FastGetRequest, reply *GetReply) {
	s.mu.Lock()

	var l uint64
	l = 0
	sid := shardOf(args.Key)
	if s.shardMap[sid] == true {
		reply.Value = s.kvss[sid][args.Key].val
		l = s.kvss[sid][args.Key].ver
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
		return
	}
	s.mu.Unlock()
	s.opLog.WaitAppend(l)
}

// TODO: InstallShard()

func RecoverGoKVShardServer() *GoKVShardServer {
	// want to go through the entire operation log
	return nil
}

func MakeGoKVShardServer() *GoKVShardServer {
	srv := new(GoKVShardServer)
	srv.mu = new(sync.Mutex)
	srv.lastReply = make(map[uint64]GetReply)
	srv.lastSeq = make(map[uint64]uint64)
	srv.kvss = make(map[uint64]map[uint64]VersionedValue)
	srv.shardMap = make(map[uint64]bool)
	srv.opLog = aof.CreateAppendOnlyFile("redis-db/kvdur_log")
	return srv
}

func (gkv *GoKVShardServer) Start() {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[KV_PUT] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(PutReply)
		gkv.PutRPC(decodePutRequest(rawReq), rep)
		*rawReply = encodePutReply(rep)
	}

	handlers[KV_GET] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(GetReply)
		gkv.GetRPC(decodeGetRequest(rawReq), rep)
		*rawReply = encodeGetReply(rep)
	}

	handlers[KV_GET] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(GetReply)
		*rawReply = encodeGetReply(rep)
		return
		gkv.FastUnsafeGetRPC(decodeFastGetRequest(rawReq), rep)
	}
	s := rpc.MakeRPCServer(handlers)
	go s.Serve(":12345")
}
