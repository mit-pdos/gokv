package mpaxos

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type ReconnectingClient struct {
	mu        *sync.Mutex
	valid     bool
	urpcCl        *urpc.Client
	making    bool
	made_cond *sync.Cond
	addr      grove_ffi.Address
}

func MakeReconnectingClient() {
}

func (cl *ReconnectingClient) getClient() *urpc.Client {
	cl.mu.Lock()
	if cl.valid {
		ret := cl.urpcCl
		cl.mu.Unlock()
		return ret
	}

	// otherwise, make a new client
	cl.making = true
	cl.mu.Unlock()
	newRpcCl := urpc.MakeClient(cl.addr)

	cl.mu.Lock()
	cl.urpcCl = newRpcCl
	cl.made_cond.Broadcast()
	cl.valid = true
	cl.making = false
	cl.mu.Unlock()
	return newRpcCl
}

func (cl *ReconnectingClient) Call(rpcid uint64, args []byte, reply *[]byte, timeout_ms uint64) uint64 {
	urpcCl := cl.getClient()
	return urpcCl.Call(rpcid, args, reply, timeout_ms)
}
