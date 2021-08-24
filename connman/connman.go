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
	making map[HostName]*sync.Cond // a key exists iff someone is making the RPCClient for that host right now
}

func (c *ConnMan) getNewClient(host HostName) *rpc.RPCClient {
	// want to open a new RPCClient without a thundering herd of threads all
	// making their own RPCClient
	var cl *rpc.RPCClient

	// XXX: This is written under the assumption that MakeRPCClient
	// takes a long time compared to the other critical sections of c.mu
	// (e.g. this might establish a TCP connection, incurring some
	// network delay)
	cond, ok := c.making[host];
	if ok { // someone else is making the host
		cond.Wait()
		cl = c.rpcCls[host]
	} else {
		c.making[host] = sync.NewCond(c.mu)
		c.mu.Unlock()
		cl = rpc.MakeRPCClient(host)
		c.mu.Lock()
		c.rpcCls[host] = cl
		c.making[host].Broadcast()
		delete(c.making, host)
	}
	return cl
}

// This repeatedly retries the RPC after retryTimeout until it gets a response.
func (c *ConnMan) CallAtLeastOnce(host HostName, rpcid uint64, args []byte, reply *[]byte, retryTimeout uint64) {
	var cl *rpc.RPCClient
	c.mu.Lock()
	cl, ok := c.rpcCls[host]
	if !ok {
		cl = c.getNewClient(host)
	}
	c.mu.Unlock()

	for {
		err := cl.Call(rpcid, args, reply, retryTimeout)
		if err == rpc.ErrTimeout {
			// just retry
			continue
		} else if err == rpc.ErrDisconnect {
			// need to try reconnecting
			c.mu.Lock()
			if cl != c.rpcCls[host] { // our RPCClient is already out of date
				cl = c.rpcCls[host]
			} else {
				delete(c.rpcCls, host)
				cl = c.getNewClient(host)
			}
			c.mu.Unlock()
			continue
		} else {
			break
		}
	}
}
