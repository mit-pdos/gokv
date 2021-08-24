package connman

// Provides a connection manager, which allows one to make RPCs against any
// hosts while using only one underlying network connection to each host; this
// also tries reconnnecting on failures.

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"sync"
)

type HostName = rpc.HostName

type ConnMan struct {
	mu     *sync.Mutex
	rpcCls map[HostName]*rpc.RPCClient
	making map[HostName]*sync.Cond
}

// This repeatedly retries the RPC after retryTimeout until it gets a response.
func (c *ConnMan) CallAtLeastOnce(host HostName, rpcid uint64, args []byte, reply *[]byte, retryTimeout uint64) {
	c.mu.Lock()
	cl := c.rpcCls[host]
	c.mu.Unlock()
	for {
		err := cl.Call(rpcid, args, reply, retryTimeout)
		if err == rpc.ErrTimeout {
			// just retry
			continue
		} else if err == rpc.ErrDisconnect {
			// need to try reconnecting, but want to avoid thundering herd in
			// case a bunch of threads are using the same RPCClient
			c.mu.Lock()
			// XXX: This is written under the assumption that MakeRPCClient
			// takes a long time compared to the other critical sections of c.mu
			// (e.g. this might establish a TCP connection, incurring some
			// network delay)
			cond, ok := c.making[host];
			if ok {
				cond.Wait()
				cl = c.rpcCls[host]
				c.mu.Unlock()
			} else {
				c.making[host] = sync.NewCond(c.mu)
				c.mu.Unlock()
				cl = rpc.MakeRPCClient(host)
				c.mu.Lock()
				c.rpcCls[host] = cl
				c.making[host].Broadcast()
				delete(c.making, host)
			}
			continue
		} else {
			break
		}
	}
}
