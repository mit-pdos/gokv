package goosekv

import (
	"github.com/tchajed/marshal"
	"github.com/mit-pdos/gokv/aof"
	"github.com/mit-pdos/lockservice"
	"sync"
)

type ValueType = uint64

type GoKVServer struct {
	mu        *sync.Mutex
	lastReply map[uint64]uint64
	lastSeq   map[uint64]uint64
	kvs       map[uint64]ValueType
	opLog     *aof.AppendOnlyFile
}

type PutArgs struct {
	Key   uint64
	Value ValueType
}

func encodeReq(optype uint64, args *lockservice.RPCRequest) []byte {
	// num_bytes := uint64(8 + 8 + len(args.Value)) // key + value-len + value
	num_bytes := uint64(8 + 8 + 8 + 8 + 8) // ReqType + CID + Seq + key + value
	e := marshal.NewEnc(num_bytes)
	e.PutInt(optype)
	e.PutInt(args.CID)
	e.PutInt(args.Seq)
	e.PutInt(args.Args.U64_1)
	e.PutInt(args.Args.U64_2)

	// TODO: support []byte args
	// e.PutInt(uint64(len(args.Value)))
	// e.PutBytes(args.Value)

	return e.Finish()
}

func (s *GoKVServer) Put(args *lockservice.RPCRequest, reply *lockservice.RPCReply) bool {
	s.mu.Lock()
	if lockservice.CheckReplyTable(s.lastSeq, s.lastReply, args.CID, args.Seq, reply) {
	} else {
		s.kvs[args.Args.U64_1] = args.Args.U64_2
	}
	l := s.opLog.Append(encodeReq(0, args))
	s.mu.Unlock()
	s.opLog.WaitAppend(l)
	return false
}

func (s *GoKVServer) Get(args *lockservice.RPCRequest, reply *lockservice.RPCReply) bool {
	s.mu.Lock()
	if lockservice.CheckReplyTable(s.lastSeq, s.lastReply, args.CID, args.Seq, reply) {
	} else {
		reply.Ret = s.kvs[args.Args.U64_1]
	}
	l := s.opLog.Append(encodeReq(1, args))
	s.mu.Unlock()

	s.opLog.WaitAppend(l)
	return false
}

// TODO: recovery

func MakeGoKVServer() *GoKVServer {
	srv := new(GoKVServer)
	srv.mu = new(sync.Mutex)
	srv.lastReply = make(map[uint64]uint64)
	srv.lastSeq = make(map[uint64]uint64)
	srv.kvs = make(map[uint64]ValueType)
	srv.opLog = aof.CreateAppendOnlyFile("kvdur_log")
	return srv
}
