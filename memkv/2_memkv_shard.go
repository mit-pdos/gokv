package memkv

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
	"log"
)

type KvMap = map[uint64][]byte

type MemKVShardServer struct {
	me        string //
	mu        *sync.Mutex
	lastReply map[uint64]GetReply
	lastSeq   map[uint64]uint64
	nextCID   uint64 // next CID that can be granted to a client

	shardMap []bool // \box(size=NSHARDS)
	// if anything is in shardMap, then we have a map[] initialized in kvss
	kvss  []KvMap // \box(size=NSHARDS)
	peers map[HostName]*MemKVShardClerk
}

type PutArgs struct {
	Key   uint64
	Value ValueType
}

func (s *MemKVShardServer) put_inner(args *PutRequest, reply *PutReply) {
	last, ok := s.lastSeq[args.CID]
	seq := args.Seq
	if ok && seq <= last {
		// XXX: this is a bit hacky
		reply.Err = s.lastReply[args.CID].Err
		return
	}
	s.lastSeq[args.CID] = args.Seq

	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		s.kvss[sid][args.Key] = args.Value // give ownership of the slice to the server
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
	}

	// XXX: this is a bit hacky (same as above)
	s.lastReply[args.CID] = GetReply{Err: reply.Err}
}

func (s *MemKVShardServer) PutRPC(args *PutRequest, reply *PutReply) {
	s.mu.Lock()
	s.put_inner(args, reply)
	s.mu.Unlock()
}

func (s *MemKVShardServer) get_inner(args *GetRequest, reply *GetReply) {
	last, ok := s.lastSeq[args.CID]
	seq := args.Seq
	if ok && seq <= last {
		*reply = s.lastReply[args.CID]
		return
	}
	s.lastSeq[args.CID] = args.Seq

	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		reply.Value = append(make([]byte, 0), s.kvss[sid][args.Key]...)
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
	}
	s.lastReply[args.CID] = *reply
}

func (s *MemKVShardServer) GetRPC(args *GetRequest, reply *GetReply) {
	s.mu.Lock()
	s.get_inner(args, reply)
	s.mu.Unlock()
}

// NOTE: easy to do a little optimization with shard migration:
// add a "RemoveShard" rpc, which removes the shard on the target server, and
// returns half of the ghost state for that shard. Meanwhile, InstallShard()
// will only grant half the ghost state, and physical state will keep track of
// the fact that the shard is only good for read-only operations up until that
// flag is updated (i.e. until RemoveShard() is run).
func (s *MemKVShardServer) install_shard_inner(args *InstallShardRequest) {
	last, ok := s.lastSeq[args.CID]
	seq := args.Seq
	if ok && seq <= last {
		return
	}
	s.lastSeq[args.CID] = args.Seq

	s.shardMap[args.Sid] = true
	s.kvss[args.Sid] = args.Kvs
	s.lastReply[args.CID] = GetReply{Err: 0, Value: nil}
}

func (s *MemKVShardServer) InstallShardRPC(args *InstallShardRequest) {
	s.mu.Lock()
	s.install_shard_inner(args)
	s.mu.Unlock()
}

func (s *MemKVShardServer) MoveShardRPC(args *MoveShardRequest) {
	s.mu.Lock()
	_, ok := s.peers[args.Dst]
	if !ok {
		// s.mu.Unlock()
		ck := MakeFreshKVClerk(args.Dst)
		// s.mu.Lock()
		s.peers[args.Dst] = ck
	}

	if !s.shardMap[args.Sid] {
		s.mu.Unlock()
		return
	}
	kvs := s.kvss[args.Sid]
	s.kvss[args.Sid] = make(KvMap)
	s.shardMap[args.Sid] = false
	s.peers[args.Dst].InstallShard(args.Sid, kvs) // XXX: if we want to do this without the lock, need a lock in the clerk itself
	s.mu.Unlock()
}

func MakeMemKVShardServer() *MemKVShardServer {
	srv := new(MemKVShardServer)
	srv.mu = new(sync.Mutex)
	srv.lastReply = make(map[uint64]GetReply)
	srv.lastSeq = make(map[uint64]uint64)
	srv.shardMap = make([]bool, NSHARD)
	srv.kvss = make([]KvMap, NSHARD)
	srv.peers = make(map[HostName]*MemKVShardClerk)
	for i := uint64(0); i < NSHARD; i++ {
		srv.shardMap[i] = true
		srv.kvss[i] = make(map[uint64][]byte)
	}
	return srv
}

func (s *MemKVShardServer) GetCIDRPC() uint64 {
	log.Println("GetCIDRPC() starting")
	s.mu.Lock()
	r := s.nextCID
	s.nextCID = s.nextCID + 1
	s.mu.Unlock()
	log.Println("GetCIDRPC() done")
	return r
}

func (mkv *MemKVShardServer) Start(host string) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[KV_FRESHCID] = func(rawReq []byte, rawReply *[]byte) {
		*rawReply = encodeCID(mkv.GetCIDRPC())
	}

	handlers[KV_PUT] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(PutReply)
		mkv.PutRPC(decodePutRequest(rawReq), rep)
		*rawReply = encodePutReply(rep)
	}

	handlers[KV_GET] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(GetReply)
		mkv.GetRPC(decodeGetRequest(rawReq), rep)
		*rawReply = encodeGetReply(rep)
	}

	handlers[KV_INS_SHARD] = func(rawReq []byte, rawReply *[]byte) {
		// NOTE: decoding, i.e. construction of in-memory map, happens before we get
		// the lock
		mkv.InstallShardRPC(decodeInstallShardRequest(rawReq))
		*rawReply = make([]byte, 0)
	}

	handlers[KV_MOV_SHARD] = func(rawReq []byte, rawReply *[]byte) {
		mkv.MoveShardRPC(decodeMoveShardRequest(rawReq))
		*rawReply = make([]byte, 0)
	}
	s := rpc.MakeRPCServer(handlers)
	s.Serve(host, 1)
}
