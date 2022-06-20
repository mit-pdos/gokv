package example

import (
	rsm "github.com/mit-pdos/gokv/reconfig/replica"
	"github.com/tchajed/marshal"
)

type CtrServer struct {
	s   *rsm.Server
	ctr uint64
}

// Returns the fetched value if successful, replies with an empty response if
// unsuccessful.
func (cs *CtrServer) FetchAndIncrement(args []byte, reply *[]byte) {
	// FIXME: the underlying interface has changed, so this needs to change too
	err := cs.s.AppendOp(args)
	if err != rsm.ENone {
		*reply = make([]byte, 0)
	}
}

func (cs *CtrServer) apply(op []byte) []byte {
	ret := cs.ctr
	cs.ctr += 1
	enc := marshal.NewEnc(8)
	enc.PutInt(ret)
	return enc.Finish()
}

func (cs *CtrServer) getState() []byte {
	enc := marshal.NewEnc(8)
	enc.PutInt(cs.ctr)
	return enc.Finish()
}

func (cs *CtrServer) setState(enc_state []byte) {
	dec := marshal.NewDec(enc_state)
	cs.ctr = dec.GetInt()
}

func StartCtrServer() {
	cs := new(CtrServer)
	cs.ctr = 0
	// cs.s = rsm.MakeServer(cs.apply, cs.getState)
}
