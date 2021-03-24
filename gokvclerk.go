package gokv

import (
	"net/rpc"
)

type GoKVClerk struct {
	// mu *sync.Mutex
	// seq uint64
	// cid uint64
	Cl *rpc.Client
}

func MakeKVClerk(cid uint64, serverAddress string) *GoKVClerk {
	ck := new(GoKVClerk)
	var err error
	ck.Cl, err = rpc.DialHTTP("tcp", serverAddress + ":12345")
	if err != nil {
		panic(err)
	}
	return ck
}

func (ck *GoKVClerk) Put(key uint64, value string) {
	err := ck.Cl.Call("KV.PutRPC", &PutArgs{key, value}, &struct{}{})
	if err != nil {
		panic(err)
	}
}

func (ck *GoKVClerk) Get(key uint64) string {
	var ret string
	err := ck.Cl.Call("KV.GetRPC", &key, &ret)
	if err != nil {
		panic(err)
	}
	return ret
}
