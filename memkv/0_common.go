package memkv

import (
	"github.com/goose-lang/std"
	"github.com/tchajed/marshal"
)

type HostName = uint64

type ValueType = uint64

type ErrorType = uint64

const (
	ENone          = uint64(0)
	EDontHaveShard = uint64(1)
)

const NSHARD = uint64(65536)

// rpc ids
const KV_FRESHCID = uint64(0)
const KV_PUT = uint64(1)
const KV_GET = uint64(2)
const KV_CONDITIONAL_PUT = uint64(3)
const KV_INS_SHARD = uint64(4)
const KV_MOV_SHARD = uint64(5)

func shardOf(key uint64) uint64 {
	return key % NSHARD
}

// "universal" reply type for the reply table
type ShardReply struct {
	Err     ErrorType
	Value   []byte
	Success bool
}

type PutRequest struct {
	CID   uint64
	Seq   uint64
	Key   uint64
	Value []byte
}

// doesn't include the operation type
func EncodePutRequest(args *PutRequest) []byte {
	// assume no overflow (args.Value would have to be almost 2^64 bytes large...)
	num_bytes := std.SumAssumeNoOverflow(8+8+8+8, uint64(len(args.Value))) // CID + Seq + key + value-len + value
	e := marshal.NewEnc(num_bytes)
	e.PutInt(args.CID)
	e.PutInt(args.Seq)
	e.PutInt(args.Key)
	e.PutInt(uint64(len(args.Value)))
	e.PutBytes(args.Value)

	return e.Finish()
}

func DecodePutRequest(reqData []byte) *PutRequest {
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

func EncodePutReply(reply *PutReply) []byte {
	e := marshal.NewEnc(8)
	e.PutInt(reply.Err)
	return e.Finish()
}

func DecodePutReply(replyData []byte) *PutReply {
	reply := new(PutReply)
	d := marshal.NewDec(replyData)
	reply.Err = d.GetInt()
	return reply
}

type GetRequest struct {
	CID uint64
	Seq uint64
	Key uint64
}

type GetReply struct {
	Err   ErrorType
	Value []byte
}

func EncodeGetRequest(req *GetRequest) []byte {
	e := marshal.NewEnc(3 * 8)
	e.PutInt(req.CID)
	e.PutInt(req.Seq)
	e.PutInt(req.Key)
	return e.Finish()
}

func DecodeGetRequest(rawReq []byte) *GetRequest {
	req := new(GetRequest)
	d := marshal.NewDec(rawReq)
	req.CID = d.GetInt()
	req.Seq = d.GetInt()
	req.Key = d.GetInt()
	return req
}

func EncodeGetReply(rep *GetReply) []byte {
	// assume no overflow (rep.Value would have to be almost 2^64 bytes large...)
	num_bytes := std.SumAssumeNoOverflow(8+8, uint64(len(rep.Value))) // CID + Seq + key + value-len + value
	e := marshal.NewEnc(num_bytes)
	e.PutInt(rep.Err)
	e.PutInt(uint64(len(rep.Value)))
	e.PutBytes(rep.Value)
	return e.Finish()
}

func DecodeGetReply(rawRep []byte) *GetReply {
	rep := new(GetReply)
	d := marshal.NewDec(rawRep)
	rep.Err = d.GetInt()
	rep.Value = d.GetBytes(d.GetInt())

	return rep
}

type ConditionalPutRequest struct {
	CID           uint64
	Seq           uint64
	Key           uint64
	ExpectedValue []byte
	NewValue      []byte
}

type ConditionalPutReply struct {
	Err     ErrorType
	Success bool
}

func EncodeConditionalPutRequest(req *ConditionalPutRequest) []byte {
	// assume no overflow (req.NewValue and req.ExpectedValue together would have to be almost 2^64 bytes large...)
	// CID + Seq + key + exp-value-len + exp-value + new-value-len + new-value
	num_bytes := std.SumAssumeNoOverflow(8+8+8+8+8, std.SumAssumeNoOverflow(uint64(len(req.ExpectedValue)), uint64(len(req.NewValue))))
	e := marshal.NewEnc(num_bytes)
	e.PutInt(req.CID)
	e.PutInt(req.Seq)
	e.PutInt(req.Key)
	e.PutInt(uint64(len(req.ExpectedValue)))
	e.PutBytes(req.ExpectedValue)
	e.PutInt(uint64(len(req.NewValue)))
	e.PutBytes(req.NewValue)
	return e.Finish()
}

func DecodeConditionalPutRequest(rawReq []byte) *ConditionalPutRequest {
	req := new(ConditionalPutRequest)
	d := marshal.NewDec(rawReq)
	req.CID = d.GetInt()
	req.Seq = d.GetInt()
	req.Key = d.GetInt()
	req.ExpectedValue = d.GetBytes(d.GetInt())
	req.NewValue = d.GetBytes(d.GetInt())
	return req
}

func EncodeConditionalPutReply(reply *ConditionalPutReply) []byte {
	e := marshal.NewEnc(8 + 1)
	e.PutInt(reply.Err)
	e.PutBool(reply.Success)
	return e.Finish()
}

func DecodeConditionalPutReply(replyData []byte) *ConditionalPutReply {
	reply := new(ConditionalPutReply)
	d := marshal.NewDec(replyData)
	reply.Err = d.GetInt()
	reply.Success = d.GetBool()
	return reply
}

type InstallShardRequest struct {
	CID uint64
	Seq uint64
	Sid uint64
	Kvs map[uint64][]byte
}

// NOTE: probably can just amortize this by keeping track of this with the map itself
func SizeOfMarshalledMap(m map[uint64][]byte) uint64 {
	var s uint64
	s = 8
	for _, value := range m {
		v := std.SumAssumeNoOverflow(uint64(len(value)), 8+8)
		s = std.SumAssumeNoOverflow(s, v)
	}
	return s
}

func EncSliceMap(e marshal.Enc, m map[uint64][]byte) {
	e.PutInt(uint64(len(m)))
	for key, value := range m {
		e.PutInt(key)
		e.PutInt(uint64(len(value)))
		e.PutBytes(value)
	}
}

func DecSliceMap(d marshal.Dec) map[uint64][]byte {
	sz := d.GetInt()
	m := make(map[uint64][]byte)
	var i = uint64(0)
	for i < sz {
		k := d.GetInt()
		v := d.GetBytes(d.GetInt())
		m[k] = v
		i = i + 1
	}
	return m
}

func encodeInstallShardRequest(req *InstallShardRequest) []byte {
	num_bytes := std.SumAssumeNoOverflow(8+8+8, SizeOfMarshalledMap(req.Kvs))
	e := marshal.NewEnc(num_bytes)
	e.PutInt(req.CID)
	e.PutInt(req.Seq)
	e.PutInt(req.Sid)
	EncSliceMap(e, req.Kvs)
	return e.Finish()
}

func decodeInstallShardRequest(rawReq []byte) *InstallShardRequest {
	d := marshal.NewDec(rawReq)
	req := new(InstallShardRequest)
	req.CID = d.GetInt()
	req.Seq = d.GetInt()
	req.Sid = d.GetInt()
	req.Kvs = DecSliceMap(d)
	return req
}

type MoveShardRequest struct {
	Sid uint64
	Dst HostName
}

func encodeMoveShardRequest(req *MoveShardRequest) []byte {
	e := marshal.NewEnc(8 + 8)
	e.PutInt(req.Sid)
	e.PutInt(req.Dst)
	return e.Finish()
}

func decodeMoveShardRequest(rawReq []byte) *MoveShardRequest {
	req := new(MoveShardRequest)
	d := marshal.NewDec(rawReq)
	req.Sid = d.GetInt()
	req.Dst = d.GetInt()
	return req
}

// FIXME: these should just be in goose std or something
func EncodeUint64(i uint64) []byte {
	e := marshal.NewEnc(8)
	e.PutInt(i)
	return e.Finish()
}

func DecodeUint64(raw []byte) uint64 {
	return marshal.NewDec(raw).GetInt()
}

func encodeShardMap(shardMap *[]HostName) []byte {
	// requires that shardMap is a list
	e := marshal.NewEnc(8 * NSHARD)
	e.PutInts(*shardMap)
	return e.Finish()
}

func decodeShardMap(raw []byte) []HostName {
	d := marshal.NewDec(raw)
	return d.GetInts(NSHARD)
}
