package map_marshal

import "github.com/tchajed/marshal"

func EncodeMapU64ToU64(kvs map[uint64]uint64) []byte {
	var enc = make([]byte, 0)
	enc = marshal.WriteInt(enc, uint64(len(kvs)))
	for k, v := range kvs {
		enc = marshal.WriteInt(enc, k)
		enc = marshal.WriteInt(enc, v)
	}
	return enc
}

func DecodeMapU64ToU64(enc_in []byte) (map[uint64]uint64, []byte) {
	var enc = enc_in
	kvs := make(map[uint64]uint64, 0)
	numEntries, enc2 := marshal.ReadInt(enc)
	enc = enc2
	for i := uint64(0); i < numEntries; i++ {
		var key uint64
		var val uint64
		key, enc = marshal.ReadInt(enc)
		val, enc = marshal.ReadInt(enc)

		kvs[key] = val
	}
	return kvs, enc
}

func EncodeMapU64ToBytes(kvs map[uint64][]byte) []byte {
	var enc = make([]byte, 0)
	enc = marshal.WriteInt(enc, uint64(len(kvs)))
	for k, v := range kvs {
		enc = marshal.WriteInt(enc, k)
		enc = marshal.WriteInt(enc, uint64(len(v)))
		enc = marshal.WriteBytes(enc, v)
	}
	return enc
}

func DecodeMapU64ToBytes(enc_in []byte) (map[uint64][]byte, []byte) {
	var enc = enc_in
	kvs := make(map[uint64][]byte, 0)
	numEntries, enc2 := marshal.ReadInt(enc)
	enc = enc2
	for i := uint64(0); i < numEntries; i++ {
		key, enc3 := marshal.ReadInt(enc)
		valLen, enc4 := marshal.ReadInt(enc3)

		val, enc5 := marshal.ReadBytesCopy(enc4, valLen)
		enc = enc5

		kvs[key] = val
	}
	return kvs, enc
}
