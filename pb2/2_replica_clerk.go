package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
)

const BACKUP_APPEND = uint64(0)
const PRIMARY_GETLOG = uint64(1)
const PRIMARY_ADDREPLICA = uint64(2)

type AppendArgs struct {
	cn        uint64
	log       []LogEntry
	commitIdx uint64
}

type BackupClerk struct {
	cl *rpc.RPCClient
}

func EncodeAppendArgs(args *AppendArgs) []byte {
	var length = uint64(16)

	// NOTE: encoder shouldn't *need* to know the exact length up front...
	for _, e := range args.log {
		length += uint64(len(e)) + 8
	}
	enc := marshal.NewEnc(length)
	enc.PutInt(args.cn)
	enc.PutInt(args.commitIdx)
	for _, e := range args.log {
		enc.PutInt(uint64(len(e)))
		enc.PutBytes(e)
	}
	return enc.Finish()
}

func DecodeAppendArgs(data []byte) *AppendArgs {
	a := new(AppendArgs)
	dec := marshal.NewDec(data)
	a.cn = dec.GetInt()
	a.commitIdx = dec.GetInt()
	a.log = make([]LogEntry, 0)
	j := uint64(16)
	for j < uint64(len(data)) {
		e := dec.GetBytes(dec.GetInt())
		j += uint64(len(e)) + 8
		a.log = append(a.log, e)
	}
	return a
}

func (ck *BackupClerk) AppendRPC(args AppendArgs) bool {
	return false
}

type PrimaryClerk struct {
	cl *rpc.RPCClient
}

func (ck *PrimaryClerk) GetLogRPC() {
}
