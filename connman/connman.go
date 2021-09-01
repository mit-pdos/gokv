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

func (c *ConnMan) getClient(host HostName) *rpc.RPCClient {
	var ret *rpc.RPCClient

	c.mu.Lock()
	for {
		cl, ok := c.rpcCls[host]
		if ok {
			ret = cl
			break
		} else {
			// want to open a new RPCClient without a thundering herd of threads all
			// making their own RPCClient
			// XXX: This is written under the assumption that MakeRPCClient
			// takes a long time compared to the other critical sections of c.mu
			// (e.g. this might establish a TCP connection, incurring some
			// network delay)
			cond, ok := c.making[host];
			if ok { // someone else is making the host
				cond.Wait()
				continue
			} else {
				c.making[host] = sync.NewCond(c.mu)
				c.mu.Unlock()
				ret = rpc.MakeRPCClient(host)
				c.mu.Lock()
				c.rpcCls[host] = ret
				c.making[host].Broadcast()
				delete(c.making, host)
				break
			}
		}
	}
	c.mu.Unlock()
	return ret
}

// This repeatedly retries the RPC after retryTimeout until it gets a response.
func (c *ConnMan) CallAtLeastOnce(host HostName, rpcid uint64, args []byte, reply *[]byte, retryTimeout uint64) {
	var cl *rpc.RPCClient
	cl = c.getClient(host)

	for {
		err := cl.Call(rpcid, args, reply, retryTimeout)
		if err == rpc.ErrTimeout {
			// just retry
			continue
		} else if err == rpc.ErrDisconnect {
			// need to try reconnecting
			c.mu.Lock()
			if cl == c.rpcCls[host] { // our RPCClient might already be out of date
				delete(c.rpcCls, host)
			}
			c.mu.Unlock()
			cl = c.getClient(host)
			continue
		} else {
			break
		}
	}
}
