package memkv

import (
	"sync"

	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/connman"
	"github.com/mit-pdos/gokv/erpc"
	"github.com/mit-pdos/gokv/memkv/conditionalputreply_gk"
	"github.com/mit-pdos/gokv/memkv/conditionalputrequest_gk"
	"github.com/mit-pdos/gokv/memkv/error_gk"
	"github.com/mit-pdos/gokv/memkv/getreply_gk"
	"github.com/mit-pdos/gokv/memkv/getrequest_gk"
	"github.com/mit-pdos/gokv/memkv/kvop_gk"
	"github.com/mit-pdos/gokv/memkv/moveshardrequest_gk"
	"github.com/mit-pdos/gokv/memkv/putreply_gk"
	"github.com/mit-pdos/gokv/memkv/putrequest_gk"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/marshal"
)

type KvMap = map[uint64][]byte

type KVShardServer struct {
	me   string //
	mu   *sync.Mutex
	erpc *erpc.Server

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

func (s *KVShardServer) put_inner(args *putrequest_gk.S, reply *putreply_gk.S) {
	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		s.kvss[sid][args.Key] = args.Value // give ownership of the slice to the server
		reply.Err = error_gk.ENone
	} else {
		reply.Err = error_gk.EDontHaveShard
	}
}

func (s *KVShardServer) PutRPC(args *putrequest_gk.S, reply *putreply_gk.S) {
	s.mu.Lock()
	s.put_inner(args, reply)
	s.mu.Unlock()
}

func (s *KVShardServer) get_inner(args *getrequest_gk.S, reply *getreply_gk.S) {
	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		reply.Value = s.kvss[sid][args.Key]
		reply.Err = error_gk.ENone
	} else {
		reply.Err = error_gk.EDontHaveShard
	}
}

func (s *KVShardServer) GetRPC(args *getrequest_gk.S, reply *getreply_gk.S) {
	s.mu.Lock()
	s.get_inner(args, reply)
	s.mu.Unlock()
}

func (s *KVShardServer) conditional_put_inner(args *conditionalputrequest_gk.S, reply *conditionalputreply_gk.S) {
	sid := shardOf(args.Key)

	if s.shardMap[sid] == true {
		m := s.kvss[sid]
		equal := std.BytesEqual(args.ExpectedValue, m[args.Key])
		if equal {
			m[args.Key] = args.NewValue // give ownership of the slice to the server
		}
		reply.Success = equal
		reply.Err = error_gk.ENone
	} else {
		reply.Err = error_gk.EDontHaveShard
	}
}

func (s *KVShardServer) ConditionalPutRPC(args *conditionalputrequest_gk.S, reply *conditionalputreply_gk.S) {
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
	s.shardMap[args.Sid] = true
	s.kvss[args.Sid] = args.Kvs
	// log.Printf("SHARD FINISHED INSTALLING %d", args.Sid)
}

func (s *KVShardServer) InstallShardRPC(args *InstallShardRequest) {
	s.mu.Lock()
	s.install_shard_inner(args)
	s.mu.Unlock()
}

func (s *KVShardServer) MoveShardRPC(args *moveshardrequest_gk.S) {
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
	srv.erpc = erpc.MakeServer()
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
	return s.erpc.GetFreshCID()
}

func (mkv *KVShardServer) Start(host HostName) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	erpc := mkv.erpc

	handlers[uint64(kvop_gk.KV_FRESHCID)] = func(rawReq []byte, rawReply *[]byte) {
		*rawReply = marshal.WriteInt(make([]byte, 8), mkv.GetCIDRPC())
	}

	// TODO: for the proofs it'd be much cleaner if marshaling (and really as much as possible)
	// was inside a separate function, rather than done inline here.
	handlers[uint64(kvop_gk.KV_PUT)] = erpc.HandleRequest(func(rawReq []byte, rawReply *[]byte) {
		rep := new(putreply_gk.S)
		req, _ := putrequest_gk.Unmarshal(rawReq)
		mkv.PutRPC(&req, rep)
		*rawReply = putreply_gk.Marshal(make([]byte, 0), *rep)
	})

	handlers[uint64(kvop_gk.KV_GET)] = erpc.HandleRequest(func(rawReq []byte, rawReply *[]byte) {
		rep := new(getreply_gk.S)
		req, _ := getrequest_gk.Unmarshal(rawReq)
		mkv.GetRPC(&req, rep)
		*rawReply = getreply_gk.Marshal(make([]byte, 0), *rep)
	})

	handlers[uint64(kvop_gk.KV_CONDITIONAL_PUT)] = erpc.HandleRequest(func(rawReq []byte, rawReply *[]byte) {
		rep := new(conditionalputreply_gk.S)
		req, _ := conditionalputrequest_gk.Unmarshal(rawReq)
		mkv.ConditionalPutRPC(&req, rep)
		*rawReply = conditionalputreply_gk.Marshal(make([]byte, 0), *rep)
	})

	handlers[uint64(kvop_gk.KV_INS_SHARD)] = erpc.HandleRequest(func(rawReq []byte, rawReply *[]byte) {
		// NOTE: decoding, i.e. construction of in-memory map, happens before we get
		// the lock (but we do hold the erpc lock already...)
		mkv.InstallShardRPC(decodeInstallShardRequest(rawReq))
		*rawReply = make([]byte, 0)
	})

	handlers[uint64(kvop_gk.KV_MOV_SHARD)] = func(rawReq []byte, rawReply *[]byte) {
		req, _ := moveshardrequest_gk.Unmarshal(rawReq)
		mkv.MoveShardRPC(&req)
		*rawReply = make([]byte, 0)
	}
	s := urpc.MakeServer(handlers)
	s.Serve(host)
}
