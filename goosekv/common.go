package goosekv

import (
	"github.com/tchajed/marshal"
)

type ValueType = uint64

type ErrorType = uint64

const (
	ENone = iota
	EDontHaveShard
)

const NSHARD = 65536

// rpc ids
const KV_PUT = 1
const KV_GET = 2
const KV_INS_SHARD = 3

func shardOf(key uint64) uint64 {
	return key % NSHARD
}

type PutRequest struct {
	CID uint64
	Seq uint64
	Key uint64
	Value []byte
}

// doesn't include the operation type
func encodePutRequest(args *PutRequest) []byte {
	num_bytes := uint64(8 + 8 + 8 + 8 + len(args.Value)) // CID + Seq + key + value-len + value
	e := marshal.NewEnc(num_bytes)
	e.PutInt(args.CID)
	e.PutInt(args.Seq)
	e.PutInt(args.Key)
	e.PutInt(uint64(len(args.Value)))
	e.PutBytes(args.Value)

	return e.Finish()
}

func decodePutRequest(reqData []byte) *PutRequest {
	req := new(PutRequest)
	d := marshal.NewDec(reqData)
	req.CID = d.GetInt()
	req.Seq = d.GetInt()
	req.Key = d.GetInt()
	req.Value = d.GetBytes(d.GetInt())

	return req
}

type PutReply struct {
	Err ErrorType
}

func encodePutReply(reply *PutReply) []byte {
	e := marshal.NewEnc(8)
	e.PutInt(reply.Err)
	return e.Finish()
}

func decodePutReply(replyData []byte) *PutReply {
	reply := new(PutReply)
	d := marshal.NewDec(replyData)
	reply.Err = d.GetInt()
	return reply
}
