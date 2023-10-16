// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package concurrency

import (
	"context"
	"errors"

	// pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	// "go.etcd.io/etcd/api/v3/mvccpb"
	v3 "github.com/mit-pdos/gokv/etcd/v3"
)

func waitDelete(ctx context.Context, client *v3.Client, key string, rev int64) error {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wr v3.WatchResponse
	wch := client.Watch(cctx, key, []v3.OpOption{v3.WithRev(rev)})
	ret := false
	wch.ForRange(func(wr v3.WatchResponse) {
		for _, ev := range wr.Events {
			if ev.Type == mvccpb_DELETE {
				ret = true
			}
		}
	})
	if ret {
		return nil
	}
	if err := wr.Err(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.New("lost watcher waiting for delete")
}

// waitDeletes efficiently waits until all keys matching the prefix and no greater
// than the create revision are deleted.
func waitDeletes(ctx context.Context, client *v3.Client, pfx string, maxCreateRev int64) error { // (*pb.ResponseHeader, error) {
	getOpts := append(v3.WithLastCreate(), v3.WithMaxCreateRev(maxCreateRev))
	for {
		resp, err := client.Get(ctx, pfx, getOpts)
		if err != nil {
			return err // return nil, err
		}
		if len(resp.Kvs) == 0 {
			return nil // return resp.Header, nil
		}
		lastKey := string(resp.Kvs[0].Key)
		if err = waitDelete(ctx, client, lastKey, resp.Header.Revision); err != nil {
			return err // return nil, err
		}
	}
}
