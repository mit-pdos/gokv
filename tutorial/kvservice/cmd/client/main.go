package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/tutorial/lockservice"
)

func main() {
	var hostStr string
	flag.StringVar(&hostStr, "host", "", "Address of lock server")
	flag.Parse()

	if len(hostStr) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf := grove_ffi.MakeAddress(hostStr)
	ck := lockservice.MakeClerk(conf)
	l := ck.Acquire()
	log.Println("Acquired lock")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		l.Release()
		log.Println("Released lock")
		os.Exit(0)
	}()
	select {}
}
