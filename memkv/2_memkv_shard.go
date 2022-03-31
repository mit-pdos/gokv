package memkv

import (
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/urpc"
	"sync"
)

type KvMap = map[uint64][]byte

type KVShardServer struct {
	me        string //
	mu        *sync.Mutex
	lastReply map[uint64]ShardReply
	lastSeq   map[uint64]uint64
	nextCID   uint64 // next CID that can be granted to a client

	shardMap []bool // \box(size=NSHARDS)
	// if anything is in shardMap, then we have a map[] initialized in kvss
	kvss  []KvMap                    // \box(size=NSHARDS)
	peers map[HostName]*KVShardClerk // FIXME use ShardClerkSet, maybe?
	cm    *connman.ConnMan
}

type PutArgs struct {
	Key   uint64
	Value ValueType
}

func (s *KVShardServer) put_inner(args *PutRequest, reply *PutReply) {
	last, ok := s.lastSeq[args.CID]
	seq := args.Seq
	if ok && seq <= last {
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

	s.lastReply[args.CID] = ShardReply{Err: reply.Err}
}

func (s *KVShardServer) PutRPC(args *PutRequest, reply *PutReply) {
	s.mu.Lock()
	s.put_inner(args, reply)
	s.mu.Unlock()
}

func (s *KVShardServer) get_inner(args *GetRequest, reply *GetReply) {
	last, ok := s.lastSeq[args.CID]
	seq := args.Seq
	if ok && seq <= last {
		lastReply := s.lastReply[args.CID]
		reply.Err = lastReply.Err
		reply.Value = lastReply.Value
		return
	}
	s.lastSeq[args.CID] = args.Seq

	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		reply.Value = s.kvss[sid][args.Key]
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
	}
	s.lastReply[args.CID] = ShardReply{Err: reply.Err, Value: reply.Value}
}

func (s *KVShardServer) GetRPC(args *GetRequest, reply *GetReply) {
	s.mu.Lock()
	s.get_inner(args, reply)
	s.mu.Unlock()
}

func (s *KVShardServer) conditional_put_inner(args *ConditionalPutRequest, reply *ConditionalPutReply) {
	last, ok := s.lastSeq[args.CID]
	seq := args.Seq
	if ok && seq <= last {
		lastReply := s.lastReply[args.CID]
		reply.Err = lastReply.Err
		reply.Success = lastReply.Success
		return
	}
	s.lastSeq[args.CID] = args.Seq

	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		m := s.kvss[sid]
		equal := std.BytesEqual(args.ExpectedValue, m[args.Key])
		if equal {
			m[args.Key] = args.NewValue // give ownership of the slice to the server
		}
		reply.Success = equal
		reply.Err = ENone
	} else {
		reply.Err = EDontHaveShard
	}

	s.lastReply[args.CID] = ShardReply{Err: reply.Err, Success: reply.Success}
}

func (s *KVShardServer) ConditionalPutRPC(args *ConditionalPutRequest, reply *ConditionalPutReply) {
	s.mu.Lock()
	s.conditional_put_inner(args, reply)
	s.mu.Unlock()
}

// NOTE: easy to do a little optimization with shard migration:
// add a "RemoveShard" rpc, which removes the shard on the target server, and
// returns half of the ghost state for that shard. Meanwhile, InstallShard()
// will only grant half the ghost state, and physical state will keep track of
// the fact that the shard is only good for read-only operations up until that
// flag is updated (i.e. until RemoveShard() is run).
func (s *KVShardServer) install_shard_inner(args *InstallShardRequest) {
	// log.Printf("SHARD INSTALLING %d", args.Sid)
	last, ok := s.lastSeq[args.CID]
	seq := args.Seq
	if ok && seq <= last {
		return
	}
	s.lastSeq[args.CID] = args.Seq

	s.shardMap[args.Sid] = true
	s.kvss[args.Sid] = args.Kvs
	s.lastReply[args.CID] = ShardReply{Err: 0, Value: nil}
	// log.Printf("SHARD FINISHED INSTALLING %d", args.Sid)
}

func (s *KVShardServer) InstallShardRPC(args *InstallShardRequest) {
	s.mu.Lock()
	s.install_shard_inner(args)
	s.mu.Unlock()
}

func (s *KVShardServer) MoveShardRPC(args *MoveShardRequest) {
	s.mu.Lock()
	_, ok := s.peers[args.Dst]
	if !ok {
		// s.mu.Unlock()
		ck := MakeFreshKVShardClerk(args.Dst, s.cm)
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
	// log.Printf("SHARD Moving %d to %d", args.Sid, args.Dst)
	s.peers[args.Dst].InstallShard(args.Sid, kvs) // XXX: if we want to do this without the lock, need a lock in the clerk itself
	// log.Printf("SHARD Moved %d to %d", args.Sid, args.Dst)
	s.mu.Unlock()
}

func MakeKVShardServer(is_init bool) *KVShardServer {
	srv := new(KVShardServer)
	srv.mu = new(sync.Mutex)
	srv.lastReply = make(map[uint64]ShardReply)
	srv.lastSeq = make(map[uint64]uint64)
	srv.shardMap = make([]bool, NSHARD)
	srv.kvss = make([]KvMap, NSHARD)
	srv.peers = make(map[HostName]*KVShardClerk)
	srv.cm = connman.MakeConnMan()
	for i := uint64(0); i < NSHARD; i++ {
		srv.shardMap[i] = is_init
		if is_init {
			srv.kvss[i] = make(map[uint64][]byte)
		}
	}
	return srv
}

func (s *KVShardServer) GetCIDRPC() uint64 {
	s.mu.Lock()
	r := s.nextCID
	// Overflowing a 64bit counter will take a while, assume it dos not happen
	std.SumAssumeNoOverflow(s.nextCID, 1)
	s.nextCID = s.nextCID + 1
	s.mu.Unlock()
	return r
}

func (mkv *KVShardServer) Start(host HostName) {
	handlers := make(map[uint64]func([]byte, *[]byte))

	handlers[KV_FRESHCID] = func(rawReq []byte, rawReply *[]byte) {
		*rawReply = EncodeUint64(mkv.GetCIDRPC())
	}

	handlers[KV_PUT] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(PutReply)
		mkv.PutRPC(DecodePutRequest(rawReq), rep)
		*rawReply = EncodePutReply(rep)
	}

	handlers[KV_GET] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(GetReply)
		mkv.GetRPC(DecodeGetRequest(rawReq), rep)
		*rawReply = EncodeGetReply(rep)
	}

	handlers[KV_CONDITIONAL_PUT] = func(rawReq []byte, rawReply *[]byte) {
		rep := new(ConditionalPutReply)
		mkv.ConditionalPutRPC(DecodeConditionalPutRequest(rawReq), rep)
		*rawReply = EncodeConditionalPutReply(rep)
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
	s := urpc.MakeServer(handlers)
	s.Serve(host)
}
