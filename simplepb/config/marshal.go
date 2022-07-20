package config

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

func EncodeConfig(config []grove_ffi.Address) []byte {
	var enc = make([]byte, 0, 8+8*uint64(len(config)))
	enc = marshal.WriteInt(enc, uint64(len(config)))
	for _, h := range config {
		enc = marshal.WriteInt(enc, h)
	}
	return enc
}

func DecodeConfig(enc_config []byte) []grove_ffi.Address {
	var enc = enc_config
	var configLen uint64
	configLen, enc = marshal.ReadInt(enc)
	config := make([]grove_ffi.Address, configLen)
	for i := range config {
		config[i], enc = marshal.ReadInt(enc)
	}
	return config
}
