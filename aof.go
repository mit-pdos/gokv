package gokv

import (
	"os"
)

type AppendableFile struct {
	f *os.File
	counter int
}

func CreateAppendableFile(fname string) *AppendableFile {
	a := new(AppendableFile)
	var err error
	a.f, err = os.Create(fname)
	a.counter = 100
	if err != nil {
		panic(err)
	}
	return a
}

// Not safe to do concurrent Append
func (a *AppendableFile) Append(data []byte) {
	_, err := a.f.Write(data)
	if err != nil {
		panic(err)
	}
	a.counter--
	a.f.Sync()
}
