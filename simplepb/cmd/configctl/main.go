package main

import (
	"flag"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/config"
	"os"
)

func main() {
	var configHost string
	flag.StringVar(&configHost, "config", "", "address of config server (e.g. 10.0.0.1:12345)")
	flag.Parse()

	if configHost == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	ck := config.MakeClerk(grove_ffi.MakeAddress(configHost))

	config := make([]grove_ffi.Address, 1)
	config[0] = grove_ffi.MakeAddress("0.0.0.0:12101")
	ck.WriteConfig(0, config)
}
