package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/simplepb/admin"
	"github.com/mit-pdos/gokv/simplepb/config"
)

func main() {
	var confStr string
	flag.StringVar(&confStr, "conf", "", "address of config server")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" init host1 [host2 ...]")
			fmt.Println(" reconfig host1 [host2 ...]")
			fmt.Println(" getconf")
			os.Exit(1)
		}
	}

	usage_assert(confStr != "")

	confHost := grove_ffi.MakeAddress(confStr)

	a := flag.Args()
	usage_assert(len(a) > 0)
	if a[0] == "init" {
		servers := make([]grove_ffi.Address, 0)
		for _, srvStr := range a[1:] {
			servers = append(servers, grove_ffi.MakeAddress(srvStr))
		}
		err := admin.InitializeSystem(confHost, servers)
		if err != 0 {
			fmt.Printf("Error %d while initializing system\n", err)
		} else {
			fmt.Printf("Initialized system\n")
		}
	} else if a[0] == "reconfig" {
		servers := make([]grove_ffi.Address, 0)
		for _, srvStr := range a[1:] {
			servers = append(servers, grove_ffi.MakeAddress(srvStr))
		}
		err := admin.EnterNewConfig(confHost, servers)
		if err != 0 {
			fmt.Printf("Failed to switch config: %d\n", err)
		} else {
			fmt.Printf("Finished switching configuration\n")
		}
	} else if a[0] == "getconf" {
		ck := config.MakeClerk(confHost)
		conf := ck.GetConfig()
		fmt.Println("Got config")

		servers := make([]string, 0)
		for _, srv := range conf {
			servers = append(servers, grove_ffi.AddressToStr(srv))
		}
		fmt.Printf("Configuration is: %v\n", servers)
	}
}
