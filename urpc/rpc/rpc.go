package rpc

import (
	"github.com/tchajed/marshal"
	"sync"
)

type RPCServer struct {
	handlers map[uint64]func([]byte, *[]byte)
}

func (srv *RPCServer) rpcHandle(rpcid uint64, seqno uint64, senderHost []byte, data []byte) {
	sender := MakeSender(string(senderHost))
	/*
	start := time.Now()
	defer func() {
		fmt.Printf("%+v\n", time.Since(start))
	}()
	*/
	replyData := make([]byte, 0)

	srv.handlers[rpcid](data, &replyData) // call the function

	e := marshal.NewEnc(8 + 8 + uint64(len(replyData)))
	e.PutInt(seqno)
	e.PutInt(uint64(len(replyData)))
	e.PutBytes(replyData)
	Send(sender, e.Finish()) // TODO: contention? should we buffer these in userspace too?
}

func MakeRPCServer(handlers map[uint64]func([]byte, *[]byte)) *RPCServer {
	return &RPCServer{handlers: handlers}
}

func (srv *RPCServer) Serve(host string, numWorkers int) {
	recv := MakeReceiver(host)
	for i := 0; i < numWorkers; i++ {
		go srv.readThread(recv)
	}
}

func (srv *RPCServer) readThread(recv *Receiver) {
	for {
		data := Receive(recv)
		d := marshal.NewDec(data)
		rpcid := d.GetInt()
		seqno := d.GetInt()
		senderLen := d.GetInt()
		sender := d.GetBytes(senderLen)
		reqLen := d.GetInt()
		req := d.GetBytes(reqLen)
		go srv.rpcHandle(rpcid, seqno, sender, req)
	}
}

type callback struct {
	reply *[]byte
	done  *bool
	cond  *sync.Cond
}

type RPCClient struct {
	mu   *sync.Mutex
	send *Sender // for requests
	me string // host for responses
	seq  uint64 // next fresh sequence number

	pending map[uint64]*callback
}

func (cl *RPCClient) replyThread() {
	recv := MakeReceiver(cl.me)
	for {
		data := Receive(recv)
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

func MakeRPCClient(host string, me string) *RPCClient {
	cl := new(RPCClient)
	cl.send = MakeSender(host)
	cl.me = me
	cl.mu = new(sync.Mutex)
	cl.seq = 1
	cl.pending = make(map[uint64]*callback)

	go cl.replyThread()
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
	me := []byte(cl.me)

	e := marshal.NewEnc(8 + 8 + (8 + uint64(len(me))) + (8 + uint64(len(args))))
	e.PutInt(rpcid)
	e.PutInt(seqno)
	// TODO: would be nice if the marshal lib had support for len+data pairs...
	e.PutInt(uint64(len(me)))
	e.PutBytes(me)
	e.PutInt(uint64(len(args)))
	e.PutBytes(args)
	reqData := e.Finish()
	// fmt.Fprintf(os.Stderr, "%+v\n", reqData)

	Send(cl.send, reqData)

	// wait for reply
	cl.mu.Lock()
	for !*cb.done {
		cb.cond.Wait()
	}
	cl.mu.Unlock()
	return false
}
