package rpc

import (
	"github.com/tchajed/marshal"
	"sync"
	"github.com/mit-pdos/gokv/dist_ffi"
)

type RPCServer struct {
	handlers map[uint64]func([]byte, *[]byte)
}

func (srv *RPCServer) rpcHandle(sender *dist_ffi.Sender, rpcid uint64, seqno uint64, data []byte) {
	/*
	start := time.Now()
	defer func() {
		fmt.Printf("%+v\n", time.Since(start))
	}()
	*/
	replyData := make([]byte, 0)

	f := srv.handlers[rpcid] // for Goose
	f(data, &replyData) // call the function

	e := marshal.NewEnc(8 + 8 + uint64(len(replyData)))
	e.PutInt(seqno)
	e.PutInt(uint64(len(replyData)))
	e.PutBytes(replyData)
	dist_ffi.Send(sender, e.Finish()) // TODO: contention? should we buffer these in userspace too?
}

func MakeRPCServer(handlers map[uint64]func([]byte, *[]byte)) *RPCServer {
	return &RPCServer{handlers: handlers}
}

func (srv *RPCServer) Serve(host string, numWorkers uint64) {
	recv := dist_ffi.Listen(host)
	for i := uint64(0); i < numWorkers; i++ {
		go func () {
			srv.readThread(recv)
		}()
	}
}

func (srv *RPCServer) readThread(recv *dist_ffi.Receiver) {
	for {
		r := dist_ffi.Receive(recv)
		if r.E {
			continue
		}
		data := r.M
		sender := r.S
		d := marshal.NewDec(data)
		rpcid := d.GetInt()
		seqno := d.GetInt()
		reqLen := d.GetInt()
		req := d.GetBytes(reqLen)
		srv.rpcHandle(sender, rpcid, seqno, req) // XXX: this could (and probably should) be in a goroutine
	}
}

type callback struct {
	reply *[]byte
	done  *bool
	cond  *sync.Cond
}

type RPCClient struct {
	mu   *sync.Mutex
	send *dist_ffi.Sender // for requests
	seq  uint64 // next fresh sequence number

	pending map[uint64]*callback
}

func (cl *RPCClient) replyThread(recv *dist_ffi.Receiver) {
	for {
		r := dist_ffi.Receive(recv)
		if r.E {
			continue
		}
		data := r.M

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
	}
}

func MakeRPCClient(host string) *RPCClient {
	cl := new(RPCClient)
	var recv *dist_ffi.Receiver
	a := dist_ffi.Connect(host)
	cl.send = a.S
	recv = a.R
	cl.mu = new(sync.Mutex)
	cl.seq = 1
	cl.pending = make(map[uint64]*callback)

	go func () {
		cl.replyThread(recv) // Goose doesn't support parameters in a go statement
	} ()
	return cl
}

func (cl *RPCClient) Call(rpcid uint64, args []byte, reply *[]byte) bool {
	cb := callback{reply: reply, done: new(bool), cond: sync.NewCond(cl.mu)}
	*cb.done = false
	cl.mu.Lock()
	seqno := cl.seq
	cl.seq = cl.seq + 1
	cl.pending[seqno] = &cb
	cl.mu.Unlock()

	e := marshal.NewEnc(8 + 8 + (8 + uint64(len(args))))
	e.PutInt(rpcid)
	e.PutInt(seqno)
	e.PutInt(uint64(len(args)))
	e.PutBytes(args)
	reqData := e.Finish()
	// fmt.Fprintf(os.Stderr, "%+v\n", reqData)

	dist_ffi.Send(cl.send, reqData)

	// wait for reply
	cl.mu.Lock()
	for !*cb.done {
		cb.cond.Wait()
	}
	cl.mu.Unlock()
	return false
}
