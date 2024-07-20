package urpc

import (
	// "log"
	"log"
	"sync"
	// "time"

	"github.com/goose-lang/goose/machine"
	"github.com/goose-lang/std"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
)

type Server struct {
	handlers map[uint64]func([]byte, *[]byte)
}

func (srv *Server) rpcHandle(conn grove_ffi.Connection, rpcid uint64, seqno uint64, data []byte) {
	replyData := new([]byte)

	f := srv.handlers[rpcid] // for Goose
	f(data, replyData)       // call the function

	data1 := make([]byte, 0, 8+len(*replyData))
	data2 := marshal.WriteInt(data1, seqno)
	data3 := marshal.WriteBytes(data2, *replyData)
	// Ignore errors (what would we do about them anyway -- client will inevitably time out, and then retry)
	grove_ffi.Send(conn, data3) // TODO: contention? should we buffer these in userspace too?
}

func MakeServer(handlers map[uint64]func([]byte, *[]byte)) *Server {
	return &Server{handlers: handlers}
}

func (srv *Server) readThread(conn grove_ffi.Connection) {
	// lastRpcTime := time.Now()
	for {
		r := grove_ffi.Receive(conn)
		if r.Err {
			// This connection is *done* -- quit the thread.
			break
		}
		data := r.Data
		rpcid, data := marshal.ReadInt(data)
		seqno, data := marshal.ReadInt(data)
		req := data // remaining data
		// thisRpcTime := time.Now()
		// if machine.RandomUint64()%1024 == 0 {
		// log.Printf("urpc time between RPCs: %v\n", thisRpcTime.Sub(lastRpcTime))
		// }
		// lastRpcTime = thisRpcTime
		go func() { srv.rpcHandle(conn, rpcid, seqno, req) }() // XXX: this could (and probably should) be in a goroutine YYY: but readThread is already its own goroutine, so that seems redundant?
		continue
	}
}

func (srv *Server) Serve(host grove_ffi.Address) {
	listener := grove_ffi.Listen(grove_ffi.Address(host))
	go func() {
		for {
			conn := grove_ffi.Accept(listener)
			go func() {
				srv.readThread(conn)
			}()
		}
	}()
}

const callbackStateWaiting uint64 = 0
const callbackStateDone uint64 = 1
const callbackStateAborted uint64 = 2

type Callback struct {
	reply *[]byte
	state *uint64
	cond  *sync.Cond
}

type Client struct {
	mu   *sync.Mutex
	conn grove_ffi.Connection // for requests
	seq  uint64               // next fresh sequence number

	pending map[uint64]*Callback
}

func (cl *Client) replyThread() {
	for {
		r := grove_ffi.Receive(cl.conn)
		if r.Err {
			// This connection is unusable, so quit the thread and wake all pending requests.
			cl.mu.Lock()
			for _, cb := range cl.pending {
				*cb.state = callbackStateAborted
				cb.cond.Signal()
			}
			cl.mu.Unlock()
			break
		}
		data := r.Data

		seqno, data := marshal.ReadInt(data)
		reply := data
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

func TryMakeClient(host_name grove_ffi.Address) (uint64, *Client) {
	host := grove_ffi.Address(host_name)
	a := grove_ffi.Connect(host)
	var nilClient *Client
	if a.Err {
		return 1, nilClient
	}

	cl := &Client{
		conn:    a.Connection,
		mu:      new(sync.Mutex),
		seq:     1,
		pending: make(map[uint64]*Callback)}

	go func() {
		cl.replyThread() // Goose doesn't support parameters in a go statement
	}()
	return 0, cl
}

func MakeClient(host_name grove_ffi.Address) *Client {
	err, cl := TryMakeClient(host_name)
	if err != 0 {
		log.Printf("Unable to connect to %s", grove_ffi.AddressToStr(host_name))
	}
	machine.Assume(err == 0)
	return cl
}

type Error = uint64

const ErrNone uint64 = 0
const ErrTimeout uint64 = 1
const ErrDisconnect uint64 = 2

func (cl *Client) CallStart(rpcid uint64, args []byte) (*Callback, Error) {
	// log.Printf("Started call %d\n", rpcid)
	reply_buf := new([]byte)
	cb := &Callback{reply: reply_buf, state: new(uint64), cond: sync.NewCond(cl.mu)}
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

	data1 := make([]byte, 0, 8+8+len(args))
	data2 := marshal.WriteInt(data1, rpcid)
	data3 := marshal.WriteInt(data2, seqno)
	reqData := marshal.WriteBytes(data3, args)
	// fmt.Fprintf(os.Stderr, "%+v\n", reqData)

	if grove_ffi.Send(cl.conn, reqData) {
		// An error occured; this client is dead.
		// (&Callback works around goose not translating "nil" properly.)
		return &Callback{}, ErrDisconnect
	}

	return cb, ErrNone
}

func (cl *Client) CallComplete(cb *Callback, reply *[]byte, timeout_ms uint64) Error {

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
		*reply = *cb.reply
		cl.mu.Unlock()
		return 0 // no error
	} else {
		cl.mu.Unlock()
		if state == callbackStateAborted {
			return ErrDisconnect
		} else {
			// FIXME: in case of timeout, resend message with the same "seq", so that if the
			// server responds to a resend, we accept that response for the original request.
			return ErrTimeout
		}
	}
}

func (cl *Client) Call(rpcid uint64, args []byte, reply *[]byte, timeout_ms uint64) Error {
	cb, err := cl.CallStart(rpcid, args)
	if err != 0 {
		return err
	}
	return cl.CallComplete(cb, reply, timeout_ms)
}
