package main

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/goose/machine"
)

// copied from grove_ffi/filesys.go and modified
func AppendWithoutSync(filename string, data []byte) {
	filename = filepath.Join("durable", filename)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	f.Write(data)
	// f.Sync()

	err = f.Close()
	if err != nil {
		panic(err)
	}
}

func bench_onesize(fname string, writeSize uint64) float64 {
	// FIXME: try making data non-zero.
	data := make([]byte, writeSize)
	warmup := uint64(100)
	n := uint64(100)
	for i := uint64(0); i < warmup; i += 1 {
		AppendWithoutSync(fname, data)
	}
	start := machine.TimeNow()

	for i := uint64(0); i < n; i += 1 {
		AppendWithoutSync(fname, data)
	}
	end := machine.TimeNow()
	numWritesPerSec := float64(n) / (float64(end-start) / 1e9)
	// numBytesPerSec = float64(writeSize*n) / float64(end-start)
	return numWritesPerSec
}

func main() {
	fname := "test.data"
	grove_ffi.FileWrite(fname, nil)
	fmt.Printf("16-byte writes -> %f writes/sec\n", bench_onesize(fname, 16))
	grove_ffi.FileWrite(fname, nil)
	fmt.Printf("32-byte writes -> %f writes/sec\n", bench_onesize(fname, 32))
	grove_ffi.FileWrite(fname, nil)
	fmt.Printf("1024-byte writes -> %f writes/sec\n", bench_onesize(fname, 1024))
	grove_ffi.FileWrite(fname, nil)
	fmt.Printf("4096-byte writes -> %f writes/sec\n", bench_onesize(fname, 4096))
	grove_ffi.FileWrite(fname, nil)
	fmt.Printf("8192-byte writes -> %f writes/sec\n", bench_onesize(fname, 8*1024))
	grove_ffi.FileWrite(fname, nil)
	fmt.Printf("16384-byte writes -> %f writes/sec\n", bench_onesize(fname, 16*1024))

	sz := uint64(0)
	for i := 0; i < 20; i += 1 {
		sz += 32 * 1024
		grove_ffi.FileWrite(fname, nil)
		fmt.Printf("%d-byte writes -> %f writes/sec\n", sz, bench_onesize(fname, sz))
	}
}
