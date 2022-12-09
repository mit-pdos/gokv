package reconnectclient

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
)

type ReconnectingClient struct {
	mu        *sync.Mutex
	valid     bool
	urpcCl    *urpc.Client
	making    bool
	made_cond *sync.Cond
	addr      grove_ffi.Address
}

func MakeReconnectingClient(addr grove_ffi.Address) *ReconnectingClient {
	r := new(ReconnectingClient)
	r.mu = new(sync.Mutex)
	r.valid = false
	r.making = false
	r.made_cond = sync.NewCond(r.mu)
	r.addr = addr
	return r
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
	var newRpcCl *urpc.Client
	for {
		var err uint64
		err, newRpcCl = urpc.TryMakeClient(cl.addr)
		if err == 0 {
			break
		} else {
			machine.Sleep(10_000_000) // 10ms
			continue
		}
	}

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
	err := urpcCl.Call(rpcid, args, reply, timeout_ms)
	if err == urpc.ErrDisconnect {
		cl.mu.Lock()
		cl.valid = false
		cl.mu.Unlock()
	}
	return err
}

func (cl *ReconnectingClient) CallStart(rpcid uint64, args []byte, reply *[]byte, timeout_ms uint64) func() uint64 {
	urpcCl := cl.getClient()
	err, cb := urpcCl.CallStart(rpcid, args)
	if err == urpc.ErrDisconnect {
		cl.mu.Lock()
		cl.valid = false
		cl.mu.Unlock()
	}
	return func() uint64 {
		if err == urpc.ErrDisconnect {
			return err
		}
		return urpcCl.CallComplete(cb, reply, timeout_ms)
	}
}
