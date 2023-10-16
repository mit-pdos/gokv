package clientv3

import (
	"context"
	"github.com/mit-pdos/gokv/etcd/fakechan"
)

type WatchChan = fakechan.Chan[WatchResponse]

type OpOption struct{}

func WithRev(rev int64) OpOption {
	panic("axiom")
}

func WithLastCreate() []OpOption {
	panic("axiom")
}

func WithMaxCreateRev(maxCreateRev int64) OpOption {
	panic("axiom")
}

type Event struct {
	Type int32
}

type WatchResponse struct {
	Events []Event
}

func (wr *WatchResponse) Err() error {
	panic("axiom")
}

type Client struct {
}

func (cl *Client) Watch(ctx context.Context, key string, ops ...[]OpOption) WatchChan {
	panic("axiom")
}

type KeyValue struct {
}

type GetResponse struct {
	Kvs []KeyValue
}

func (cl *Client) Get(ctx context.Context, key string, ops ...[]OpOption) (GetResponse, error) {
	panic("axiom")
}
