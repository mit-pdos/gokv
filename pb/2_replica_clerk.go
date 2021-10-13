package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
)

const REPLICA_APPEND = uint64(0)
const REPLICA_GETLOG = uint64(1)

// const PRIMARY_ADDREPLICA = uint64(2)

// RPC arg/reply types + marshalling code for them

type AppendArgs struct {
	cn        uint64
	commitIdx uint64
	log       []LogEntry
}

func EncodeAppendArgs(args *AppendArgs) []byte {
	enc := marshal.NewEnc(16 + uint64(len(args.log)))
	enc.PutInt(args.cn)
	enc.PutInt(args.commitIdx)
	enc.PutBytes(args.log)
	return enc.Finish()
}

func DecodeAppendArgs(raw_args []byte) *AppendArgs {
	a := new(AppendArgs)
	dec := marshal.NewDec(raw_args)
	a.cn = dec.GetInt()
	a.commitIdx = dec.GetInt()
	a.log = dec.GetBytes(uint64(len(raw_args)) - 16)
	return a
}

type BecomePrimaryArgs struct {
	cn     uint64
	conf *PBConfiguration
}

type ReplicaClerk struct {
	cl *rpc.RPCClient
}

func (ck *ReplicaClerk) AppendRPC(args *AppendArgs) bool {
	raw_args := EncodeAppendArgs(args)
	reply := new([]byte)
	err := ck.cl.Call(REPLICA_APPEND, raw_args, reply, 100 /* ms */)
	if err == 0 && len(*reply) > 0 {
		return true
	}
	return false
}

func MakeReplicaClerk(host rpc.HostName) *ReplicaClerk {
	ck := new(ReplicaClerk)
	ck.cl = rpc.MakeRPCClient(host)
	return nil
}
