package util

import (
	"github.com/mit-pdos/gokv/grove_ffi"
)

type Configuration []grove_ffi.Address

func EncodeConfiguration(c *Configuration) []byte {
	panic("util: impl")
}

func DecodeConfiguration(raw_config []byte) Configuration {
	panic("util: impl")
}
