package reconnectclient

import (
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
	"github.com/tchajed/goose/machine"
)

type ReconnectingClient struct {
	mu     *sync.Mutex
	valid  bool
	urpcCl *urpc.Client
	// making    bool
	// made_cond *sync.Cond
	addr grove_ffi.Address
}

func MakeReconnectingClient(addr grove_ffi.Address) *ReconnectingClient {
	r := new(ReconnectingClient)
	r.mu = new(sync.Mutex)
	r.valid = false
	// r.making = false
	// r.made_cond = sync.NewCond(r.mu)
	r.addr = addr
	return r
}

func (cl *ReconnectingClient) getClient() (uint64, *urpc.Client) {
	cl.mu.Lock()
	if cl.valid {
		ret := cl.urpcCl
		cl.mu.Unlock()
		return 0, ret
	}

	// otherwise, make a new client
	// cl.making = true
	cl.mu.Unlock()
	var newRpcCl *urpc.Client
	var err uint64
	err, newRpcCl = urpc.TryMakeClient(cl.addr)

	if err != 0 {
		// FIXME: get rid of this throttling, now that there's no loop?
		machine.Sleep(10_000_000) // 10ms
	}

	cl.mu.Lock()
	// cl.making = false

	if err == 0 {
		cl.urpcCl = newRpcCl
		// cl.made_cond.Broadcast()
		cl.valid = true
	}
	cl.mu.Unlock()

	return err, newRpcCl
}

func (cl *ReconnectingClient) Call(rpcid uint64, args []byte, reply *[]byte, timeout_ms uint64) uint64 {
	err1, urpcCl := cl.getClient()
	if err1 != 0 {
		return err1
	}
	err := urpcCl.Call(rpcid, args, reply, timeout_ms)
	if err == urpc.ErrDisconnect {
		cl.mu.Lock()
		cl.valid = false
		cl.mu.Unlock()
	}
	return err
}

func (cl *ReconnectingClient) CallStart(rpcid uint64, args []byte, reply *[]byte, timeout_ms uint64) func() uint64 {
	err1, urpcCl := cl.getClient()
	if err1 != 0 {
		return func() uint64 { return err1 }
	}
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
