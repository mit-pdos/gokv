// Implements "exactly-once RPCs" with a reply table.
package erpc

import (
	"sync"

	"github.com/goose-lang/std"
	"github.com/tchajed/marshal"
)

type Server struct {
	mu        *sync.Mutex
	lastSeq   map[uint64]uint64
	lastReply map[uint64][]byte
	nextCID   uint64
}

func (t *Server) HandleRequest(handler func(raw_args []byte, reply *[]byte)) func(raw_args []byte, reply *[]byte) {
	return func(raw_args []byte, reply *[]byte) {
		cid, raw_args := marshal.ReadInt(raw_args)
		seq, raw_args := marshal.ReadInt(raw_args)

		t.mu.Lock()
		// check if we've seen this request before
		// (seq is definitely not 0, so if cid is not in the map this still works)
		last := t.lastSeq[cid]
		if seq <= last {
			// Old request repeated. This is either request `last`, and we send back that reply, or an
			// even older one in which case we can send whatever since the client will already have
			// moved on.
			*reply = t.lastReply[cid]
			t.mu.Unlock()
			return
		}

		handler(raw_args, reply)

		t.lastSeq[cid] = seq
		t.lastReply[cid] = *reply
		t.mu.Unlock()
	}
}

func (t *Server) GetFreshCID() uint64 {
	t.mu.Lock()
	r := t.nextCID
	// Overflowing a 64bit counter will take a while, assume it dos not happen
	t.nextCID = std.SumAssumeNoOverflow(t.nextCID, 1)
	t.mu.Unlock()
	return r
}

func MakeServer() *Server {
	t := new(Server)
	t.lastReply = make(map[uint64][]byte)
	t.lastSeq = make(map[uint64]uint64)
	t.nextCID = 0
	t.mu = new(sync.Mutex)
	return t
}

type Client struct {
	cid     uint64
	nextSeq uint64
}

func (c *Client) NewRequest(request []byte) []byte {
	seq := c.nextSeq
	c.nextSeq = std.SumAssumeNoOverflow(c.nextSeq, 1)

	data1 := make([]byte, 0, 8+8+len(request))
	data2 := marshal.WriteInt(data1, c.cid)
	data3 := marshal.WriteInt(data2, seq)
	data4 := marshal.WriteBytes(data3, request)
	return data4
}

func MakeClient(cid uint64) *Client {
	c := new(Client)
	c.cid = cid
	// On the server, we rely on no request ever having seq 0.
	c.nextSeq = 1
	return c
}
