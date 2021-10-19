package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/pb/controller"
	"os"
)

func main() {
	var ctrlStr string
	flag.StringVar(&ctrlStr, "ctrl", "", "address of controller")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" add HOST")
			os.Exit(1)
		}
	}

	usage_assert(ctrlStr != "")

	ctrl := grove_ffi.MakeAddress(ctrlStr)
	ck := controller.MakeControllerClerk(ctrl)

	a := flag.Args()
	usage_assert(len(a) > 0)
	if a[0] == "add" {
		usage_assert(len(a) == 2)
		h := grove_ffi.MakeAddress(a[1])
		ck.AddNewServer(h)
		fmt.Printf("Tried adding %s\n", a[1])
	}
}
