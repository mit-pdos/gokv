package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
)

const REPLICA_APPEND = uint64(0)
const REPLICA_GETLOG = uint64(1)
const REPLICA_BECOMEPRIMARY = uint64(2)
const REPLICA_HEARTBEAT = uint64(3)

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
	Cn   uint64
	Conf *Configuration
}

func EncodeBecomePrimaryArgs(args *BecomePrimaryArgs) []byte {
	encodedConf := EncodePBConfiguration(args.Conf)
	enc := marshal.NewEnc(8 + uint64(len(encodedConf)))
	enc.PutInt(args.Cn)
	enc.PutBytes(encodedConf)
	return enc.Finish()
}

func DecodeBecomePrimaryArgs(raw_args []byte) *BecomePrimaryArgs {
	a := new(BecomePrimaryArgs)
	dec := marshal.NewDec(raw_args)
	a.Cn = dec.GetInt()
	a.Conf = DecodePBConfiguration(raw_args[8:])
	return a
}

type ReplicaClerk struct {
	cl *urpc.Client
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

func (ck *ReplicaClerk) BecomePrimaryRPC(args *BecomePrimaryArgs) {
	raw_args := EncodeBecomePrimaryArgs(args)
	reply := new([]byte)
	err := ck.cl.Call(REPLICA_BECOMEPRIMARY, raw_args, reply, 20000 /* ms */)
	machine.Assume(err == 0)
}

func (ck *ReplicaClerk) HeartbeatRPC() bool {
	reply := new([]byte)
	return ck.cl.Call(REPLICA_HEARTBEAT, make([]byte, 0), reply, 1000) == 0
}

func MakeReplicaClerk(host grove_ffi.Address) *ReplicaClerk {
	ck := new(ReplicaClerk)
	ck.cl = urpc.MakeClient(host)
	return ck
}
