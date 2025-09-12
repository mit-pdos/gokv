package memkv

import (
	"github.com/goose-lang/std"
	"github.com/tchajed/marshal"
)

type HostName = uint64

type ValueType = uint64

type ErrorType = uint64

const NSHARD = uint64(65536)

func shardOf(key uint64) uint64 {
	return key % NSHARD
}

type InstallShardRequest struct {
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
	num_bytes := std.SumAssumeNoOverflow(8, SizeOfMarshalledMap(req.Kvs))
	e := marshal.NewEnc(num_bytes)
	e.PutInt(req.Sid)
	EncSliceMap(e, req.Kvs)
	return e.Finish()
}

func decodeInstallShardRequest(rawReq []byte) *InstallShardRequest {
	d := marshal.NewDec(rawReq)
	req := new(InstallShardRequest)
	req.Sid = d.GetInt()
	req.Kvs = DecSliceMap(d)
	return req
}
