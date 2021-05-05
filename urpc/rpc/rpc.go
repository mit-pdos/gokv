package rpc

import (
	"github.com/mit-pdos/gokv/dist_ffi"
	"github.com/tchajed/goose/machine"
	"github.com/tchajed/marshal"
	"sync"
)

type HostName = uint64

type RPCServer struct {
	handlers map[uint64]func([]byte, *[]byte)
}

func (srv *RPCServer) rpcHandle(sender dist_ffi.Sender, rpcid uint64, seqno uint64, data []byte) {
	replyData := new([]byte)

	f := srv.handlers[rpcid] // for Goose
	f(data, replyData)       // call the function

	machine.Assume(8+8+uint64(len(*replyData)) > uint64(len(*replyData)))
	e := marshal.NewEnc(8 + 8 + uint64(len(*replyData)))
	e.PutInt(seqno)
	e.PutInt(uint64(len(*replyData)))
	e.PutBytes(*replyData)
	// Ignore errors (what would we do about them anyway -- client needs to retry)
	dist_ffi.Send(sender, e.Finish()) // TODO: contention? should we buffer these in userspace too?
}

func MakeRPCServer(handlers map[uint64]func([]byte, *[]byte)) *RPCServer {
	return &RPCServer{handlers: handlers}
}

func (srv *RPCServer) readThread(recv dist_ffi.Receiver) {
	for {
		r := dist_ffi.Receive(recv)
		if r.Err {
			continue
		}
		data := r.Data
		sender := r.Sender
		d := marshal.NewDec(data)
		rpcid := d.GetInt()
		seqno := d.GetInt()
		reqLen := d.GetInt()
		req := d.GetBytes(reqLen)
		srv.rpcHandle(sender, rpcid, seqno, req) // XXX: this could (and probably should) be in a goroutine
		continue
	}
}

func (srv *RPCServer) Serve(host HostName, numWorkers uint64) {
	recv := dist_ffi.Listen(dist_ffi.Address(host))
	for i := uint64(0); i < numWorkers; i++ {
		go func() {
			srv.readThread(recv)
		}()
	}
}

type callback struct {
	reply *[]byte
	done  *bool
	cond  *sync.Cond
}

type RPCClient struct {
	mu   *sync.Mutex
	send dist_ffi.Sender // for requests
	seq  uint64          // next fresh sequence number

	pending map[uint64]*callback
}

func (cl *RPCClient) replyThread(recv dist_ffi.Receiver) {
	for {
		r := dist_ffi.Receive(recv)
		if r.Err {
			continue
		}
		data := r.Data

		d := marshal.NewDec(data)
		seqno := d.GetInt()
		// TODO: Can we just "read the rest of the bytes"?
		replyLen := d.GetInt()
		reply := d.GetBytes(replyLen)

		cl.mu.Lock()
		cb, ok := cl.pending[seqno]
		if ok {
			delete(cl.pending, seqno)
			*cb.reply = reply
			*cb.done = true
			cb.cond.Signal()
		}
		cl.mu.Unlock()
		continue
	}
}

func MakeRPCClient(host_name HostName) *RPCClient {
	host := dist_ffi.Address(host_name)
	a := dist_ffi.Connect(host)
	// Assume no error
	machine.Assume(!a.Err)

	cl := &RPCClient{
		send:    a.Sender,
		mu:      new(sync.Mutex),
		seq:     1,
		pending: make(map[uint64]*callback)}

	go func() {
		cl.replyThread(a.Receiver) // Goose doesn't support parameters in a go statement
	}()
	return cl
}

func (cl *RPCClient) Call(rpcid uint64, args []byte, reply *[]byte) bool {
	reply_buf := new([]byte)
	cb := &callback{reply: reply_buf, done: new(bool), cond: sync.NewCond(cl.mu)}
	*cb.done = false
	cl.mu.Lock()
	seqno := cl.seq
	// Overflowing a 64bit counter will take a while, assume it does not happen
	machine.Assume(cl.seq+1 > cl.seq)
	cl.seq = cl.seq + 1
	cl.pending[seqno] = cb
	cl.mu.Unlock()

	// Assume length of args + extra bytes for header does not overflow length
	machine.Assume(8+8+(8+uint64(len(args))) > uint64(len(args)))
	e := marshal.NewEnc(8 + 8 + (8 + uint64(len(args))))
	e.PutInt(rpcid)
	e.PutInt(seqno)
	e.PutInt(uint64(len(args)))
	e.PutBytes(args)
	reqData := e.Finish()
	// fmt.Fprintf(os.Stderr, "%+v\n", reqData)

	if dist_ffi.Send(cl.send, reqData) {
		// An error occured. "dist_ffi" will try to reconnect the socket;
		// make the caller try again with that new socket.
		return true
	}

	// wait for reply
	cl.mu.Lock()
	machine.WaitTimeout(cb.cond, 100 /*ms*/) // make sure we don't get stuck waiting forever
	done := *cb.done
	cl.mu.Unlock()
	if done {
		*reply = *reply_buf
		return false // no error
	} else {
		return true // error
	}
}
