package example

import (
	pb "github.com/mit-pdos/gokv/reconfig/replica"
	"github.com/tchajed/marshal"
)

type CtrServer struct {
	s *pb.Server

	lastAppliedIndex uint64
	ctr              uint64
}

// Returns the fetched value if successful, replies with an empty response if
// unsuccessful.
func (cs *CtrServer) FetchAndIncrement(args []byte, reply *[]byte) {
	err, idx := cs.s.Propose(args)
	if err != pb.ENone {
		// tell client to try elsewhere
	}

	cs.s.GetEntry(idx)
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
