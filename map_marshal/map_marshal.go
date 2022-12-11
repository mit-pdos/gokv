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
	numEntries, enc := marshal.ReadInt(enc)
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
	numEntries, enc := marshal.ReadInt(enc)
	for i := uint64(0); i < numEntries; i++ {
		var key uint64
		var valLen uint64
		key, enc = marshal.ReadInt(enc)
		valLen, enc = marshal.ReadInt(enc)

		// XXX: this would keep the whole original encoded slice around in
		// memory. We probably don't want that, so making a copy instead.
		// val = enc[:valLen]
		val := make([]byte, valLen)
		copy(val, enc[:valLen])
		enc = enc[valLen:]

		kvs[key] = val
	}
	return kvs, enc
}
