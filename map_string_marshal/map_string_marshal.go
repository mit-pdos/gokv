package map_string_marshal

import "github.com/tchajed/marshal"

func EncodeStringMap(kvs map[string]string) []byte {
	var enc = make([]byte, 0)
	enc = marshal.WriteInt(enc, uint64(len(kvs)))
	for k, v := range kvs {
		enc = marshal.WriteInt(enc, uint64(len(k)))
		enc = marshal.WriteBytes(enc, []byte(k))
		enc = marshal.WriteInt(enc, uint64(len(v)))
		enc = marshal.WriteBytes(enc, []byte(v))
	}
	return enc
}

func DecodeStringMap(enc_in []byte) map[string]string {
	var enc = enc_in
	var numEntries uint64
	kvs := make(map[string]string, 0)

	numEntries, enc = marshal.ReadInt(enc)
	numEntries2 := numEntries
	for i := uint64(0); i < numEntries2; i++ {
		var ln uint64
		var key []byte
		var val []byte
		ln, enc = marshal.ReadInt(enc)
		key, enc = marshal.ReadBytes(enc, ln)

		ln, enc = marshal.ReadInt(enc)
		val, enc = marshal.ReadBytes(enc, ln)

		kvs[string(key)] = string(val)
	}
	return kvs
}
