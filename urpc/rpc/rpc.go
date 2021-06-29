package rpc

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
	"github.com/goose-lang/std"
	"sync"
)

type HostName = uint64

type RPCServer struct {
	handlers map[uint64]func([]byte, *[]byte)
}

func (srv *RPCServer) rpcHandle(conn grove_ffi.Connection, rpcid uint64, seqno uint64, data []byte) {
	replyData := new([]byte)

	f := srv.handlers[rpcid] // for Goose
	f(data, replyData)       // call the function

	// assume no overflow (*replyData would have to be almost 2^64 bytes large...)
	num_bytes := std.SumAssumeNoOverflow(8 + 8, uint64(len(*replyData)))
	e := marshal.NewEnc(num_bytes)
	e.PutInt(seqno)
	e.PutInt(uint64(len(*replyData)))
	e.PutBytes(*replyData)
	// Ignore errors (what would we do about them anyway -- client will inevitably time out, and then retry)
	grove_ffi.Send(conn, e.Finish()) // TODO: contention? should we buffer these in userspace too?
}

func MakeRPCServer(handlers map[uint64]func([]byte, *[]byte)) *RPCServer {
	return &RPCServer{handlers: handlers}
}

func (srv *RPCServer) readThread(conn grove_ffi.Connection) {
	for {
		r := grove_ffi.Receive(conn)
		if r.Err {
			// This connection is *done* -- quit the thread.
			break
		}
		data := r.Data
		d := marshal.NewDec(data)
		rpcid := d.GetInt()
		seqno := d.GetInt()
		reqLen := d.GetInt()
		req := d.GetBytes(reqLen)
		srv.rpcHandle(conn, rpcid, seqno, req) // XXX: this could (and probably should) be in a goroutine YYY: but readThread is already its own goroutine, so that seems redundant?
		continue
	}
}

func (srv *RPCServer) Serve(host HostName, numWorkers uint64) {
	listener := grove_ffi.Listen(grove_ffi.Address(host))
	go func() {
		for {
			conn := grove_ffi.Accept(listener);
			go func() {
				srv.readThread(conn)
			}()
		}
	}()
}

const callbackStateWaiting uint64 = 0
const callbackStateDone uint64 = 1
const callbackStateAborted uint64 = 2

type callback struct {
	reply *[]byte
	state  *uint64
	cond  *sync.Cond
}

type RPCClient struct {
	mu   *sync.Mutex
	conn grove_ffi.Connection // for requests
	seq  uint64          // next fresh sequence number

	pending map[uint64]*callback
}

func (cl *RPCClient) replyThread() {
	for {
		r := grove_ffi.Receive(cl.conn)
		if r.Err {
			// This connection is *done* -- quit the thread and wake all pending requests.
			cl.mu.Lock()
			for _, cb := range cl.pending {
				*cb.state = callbackStateAborted
				cb.cond.Signal()
			}
			cl.mu.Unlock()
			break
		}
		data := r.Data

		d := marshal.NewDec(data)
		seqno := d.GetInt()
		// TODO: Can we just "read the rest of the bytes"?
		replyLen := d.GetInt()
		reply := d.GetBytes(replyLen)
		// log.Printf("Got reply for call %d\n", seqno)

		cl.mu.Lock()
		cb, ok := cl.pending[seqno]
		if ok {
			delete(cl.pending, seqno)
			*cb.reply = reply
			*cb.state = callbackStateDone
			cb.cond.Signal()
		}
		cl.mu.Unlock()
		continue
	}
}

func MakeRPCClient(host_name HostName) *RPCClient {
	host := grove_ffi.Address(host_name)
	a := grove_ffi.Connect(host)
	// Assume no error
	// FIXME: shouldn't assume that this is error-free when we try to reconnect
	// because of a temporary network failure
	machine.Assume(!a.Err)

	cl := &RPCClient{
		conn:    a.Connection,
		mu:      new(sync.Mutex),
		seq:     1,
		pending: make(map[uint64]*callback)}

	go func() {
		cl.replyThread() // Goose doesn't support parameters in a go statement
	}()
	return cl
}

const ErrTimeout uint64 = 1
const ErrDisconnect uint64 = 2

func (cl *RPCClient) Call(rpcid uint64, args []byte, reply *[]byte, timeout_ms uint64) uint64 {
	// log.Printf("Started call %d\n", rpcid)
	reply_buf := new([]byte)
	cb := &callback{reply: reply_buf, state: new(uint64), cond: sync.NewCond(cl.mu)}
	*cb.state = callbackStateWaiting
	cl.mu.Lock()
	seqno := cl.seq
	// Overflowing a 64bit counter will take a while, assume it does not happen
	cl.seq = std.SumAssumeNoOverflow(cl.seq, 1)
	cl.pending[seqno] = cb
	cl.mu.Unlock()

	// If the `replyThread` goes down during this call, then
	// - either this happens before the above critical section,
	//   in which case the `Send` below will fail immediately;
	// - or it happens after the critical section, in which case the `replyThread` will set
	//   our status to `callbackStateAborted` which we will notice below.

	// assume no overflow (args would have to be almost 2^64 bytes large...)
	num_bytes := std.SumAssumeNoOverflow(8 + 8 + 8, uint64(len(args)))
	e := marshal.NewEnc(num_bytes)
	e.PutInt(rpcid)
	e.PutInt(seqno)
	e.PutInt(uint64(len(args)))
	e.PutBytes(args)
	reqData := e.Finish()
	// fmt.Fprintf(os.Stderr, "%+v\n", reqData)

	if grove_ffi.Send(cl.conn, reqData) {
		// An error occured; this client is dead.
		return ErrDisconnect
	}

	// wait for reply
	cl.mu.Lock()
	if *cb.state == callbackStateWaiting {
		// No reply yet (and `replyThread` hasn't aborted either).
		// Wait just a single time; Go guarantees no spurious wakeups.
		// log.Printf("Waiting for reply for call %d(%d)\n", seqno, rpcid)
		machine.WaitTimeout(cb.cond, timeout_ms) // make sure we don't get stuck waiting forever
	}
	state := *cb.state
	if state == callbackStateDone {
		*reply = *reply_buf
		cl.mu.Unlock()
		return 0 // no error
	} else {
		cl.mu.Unlock()
		if state == callbackStateAborted {
			return ErrDisconnect
		} else {
			return ErrTimeout
		}
	}
}
