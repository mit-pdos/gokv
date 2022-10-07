package connman

// Provides a connection manager, which allows one to make RPCs against any
// hosts while using only one underlying network connection to each host; this
// also tries reconnnecting on failures.

import (
	"net/rpc"
	"sync"

	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/mit-pdos/gokv/urpc"
)

type HostName = grove_ffi.Address

type ConnMan struct {
	mu     *sync.Mutex
	rpcCls map[HostName]*urpc.Client
	making map[HostName]*sync.Cond // a key exists iff someone is making the RPCClient for that host right now
}

func MakeConnMan() *ConnMan {
	c := new(ConnMan)
	c.mu = new(sync.Mutex)
	c.rpcCls = make(map[HostName]*urpc.Client)
	c.making = make(map[HostName]*sync.Cond)
	return c
}

func (c *ConnMan) getClient(host HostName) *urpc.Client {
	var ret *urpc.Client

	c.mu.Lock()
	for {
		cl, ok := c.rpcCls[host]
		if ok {
			ret = cl
			break
		}
		// want to open a new RPCClient without a thundering herd of threads all
		// making their own RPCClient
		// XXX: This is written under the assumption that MakeRPCClient
		// takes a long time compared to the other critical sections of c.mu
		// (e.g. this might establish a TCP connection, incurring some
		// network delay)
		cond, ok := c.making[host]
		if ok { // someone else is making the host
			cond.Wait()
			continue
		}
		my_cond := sync.NewCond(c.mu)
		c.making[host] = my_cond
		c.mu.Unlock()
		ret = urpc.MakeClient(host)
		c.mu.Lock()
		c.rpcCls[host] = ret
		my_cond.Broadcast()
		delete(c.making, host)
		break
	}
	c.mu.Unlock()
	return ret
}

// This repeatedly retries the RPC after retryTimeout until it gets a response.
func (c *ConnMan) CallAtLeastOnce(host HostName, rpcid uint64, args []byte, reply *[]byte, retryTimeout uint64) {
	var cl *urpc.Client
	cl = c.getClient(host)

	for {
		err := cl.Call(rpcid, args, reply, retryTimeout)
		if err == urpc.ErrTimeout {
			// just retry
			continue
		}
		if err == urpc.ErrDisconnect {
			// need to try reconnecting
			c.mu.Lock()
			if cl == c.rpcCls[host] { // our RPCClient might already be out of date
				delete(c.rpcCls, host)
			}
			c.mu.Unlock()
			cl = c.getClient(host)
			continue
		}
		break
	}
}

func (c *ConnMan) Call(host HostName, rpcid uint64, args []byte, reply *[]byte, retryTimeout uint64) rpc.ServerError {
	var cl *urpc.Client
	cl = c.getClient(host)

	err := cl.Call(rpcid, args, reply, retryTimeout)
	if err == urpc.ErrDisconnect {
		// need to try reconnecting in the future, so remove from rpcCls map
		c.mu.Lock()
		if cl == c.rpcCls[host] { // our RPCClient might already be out of date
			delete(c.rpcCls, host)
		}
		c.mu.Unlock()
		cl = c.getClient(host)
	}
	retur nerr
}
