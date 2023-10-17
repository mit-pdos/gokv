package election

type Election struct {
	lease     string
	keyPrefix string

	leaderLease string
	leaderKey   string
	leaderRev   uint64
	// hdr         *pb.ResponseHeader XXX: Not sure what clients do with this and Header()
}

type Error = uint64

const (
	ErrNone              = uint64(0)
	ErrElectionNotLeader = uint64(1)
)

type Txn struct {
}

type Op struct {
}

func OpGet(key string) *Op {
	panic("axiom")
}

func OpPutWithLease(key, val, lease string) *Op {
	panic("axiom")
}

func OpDelete(key string) *Op {
	panic("axiom")
}

func StartTxn() *Txn {
	panic("axiom")
}

func (txn *Txn) IfCreateRevisionEq(k string, ver uint64) *Txn {
	panic("axiom")
}

func (txn *Txn) Then(*Op) *Txn {
	panic("axiom")
}

func (txn *Txn) Else(*Op) *Txn {
	panic("axiom")
}

type ResponseHeader struct {
	Revision uint64
}
type KeyValue struct {
	Key            string
	Value          string
	CreateRevision uint64
}
type RangeResponse struct {
	Kvs []*KeyValue `protobuf:"bytes,2,rep,name=kvs,proto3" json:"kvs,omitempty"`
}

type ResponseOp = RangeResponse

type TxnResponse struct {
	Succeeded bool
	Header    *ResponseHeader
	Responses []*ResponseOp
}

func (txn *Txn) Commit() (*TxnResponse, Error) {
	panic("axiom")
}

func waitDeletes(pfx string, rev uint64) Error {
	panic("axiom")
}

func (e *Election) Campaign(val string) Error {

	l := e.lease
	// k := fmt.Sprintf("%s%x", e.keyPrefix, e.lease)
	k := e.keyPrefix + l
	txn := StartTxn().IfCreateRevisionEq(k, 0)
	txn = txn.Then(OpPutWithLease(k, val, e.lease))
	txn = txn.Else(OpGet(k))
	resp, err := txn.Commit()
	if err != ErrNone {
		return err
	}
	e.leaderKey, e.leaderRev, e.leaderLease = k, resp.Header.Revision, l
	if !resp.Succeeded {
		// kv := resp.Responses[0].GetResponseRange().Kvs[0]
		kv := resp.Responses[0].Kvs[0]
		e.leaderRev = kv.CreateRevision
		if string(kv.Value) != val {
			if err = e.Proclaim(val); err != ErrNone {
				e.Resign()
				return err
			}
		}
	}

	err = waitDeletes(e.keyPrefix, e.leaderRev-1)
	if err != ErrNone {
		/*
			// clean up in case of context cancel
			select {
			case <-ctx.Done():
				e.Resign(client.Ctx())
			default:
				e.leaderSession = nil
			}
			return err
		*/
	}
	// e.hdr = resp.Header

	return ErrNone
}

// Proclaim lets the leader announce a new value without another election.
func (e *Election) Proclaim(val string) Error {
	if e.leaderLease == "" {
		return ErrElectionNotLeader
	}
	txn := StartTxn().IfCreateRevisionEq(e.leaderKey, e.leaderRev)
	txn = txn.Then(OpPutWithLease(e.leaderKey, val, e.leaderLease))
	tresp, terr := txn.Commit()
	if terr != ErrNone {
		return terr
	}
	if !tresp.Succeeded {
		e.leaderKey = ""
		return ErrElectionNotLeader
	}

	// e.hdr = tresp.Header
	return ErrNone
}

// Resign lets a leader start a new election.
func (e *Election) Resign() Error {
	if e.leaderLease == "" {
		return ErrNone
	}
	// cmp := v3.Compare(v3.CreateRevision(e.leaderKey), "=", e.leaderRev)
	_, err := StartTxn().IfCreateRevisionEq(e.leaderKey, e.leaderRev).
		Then(OpDelete(e.leaderKey)).Commit()
	// if err == nil {
	// e.hdr = resp.Header
	// }
	e.leaderKey = ""
	e.leaderLease = ""
	return err
}
