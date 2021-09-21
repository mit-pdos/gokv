package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/tchajed/marshal"
	"sync"
)

type BackupClerk struct {
	cl *rpc.RPCClient
}

const BACKUP_UPDATELOG = uint64(0)

func (c *BackupClerk) UpdateLogRPC(log []byte) {
	reply := new([]byte)
	c.cl.Call(BACKUP_UPDATELOG, log, reply, 100)
}

type PrimaryServer struct {
	mu        *sync.Mutex
	opLog     []byte
	commitLog []byte
	backup    *BackupClerk
}

type BackupServer struct {
	mu        *sync.Mutex
	opLog     []byte
	commitLog []byte
	isPrimary bool
}

func (p *PrimaryServer) Add(op []byte) {
	p.mu.Lock()
	p.opLog = append(p.opLog, op...)
	o := p.opLog
	backup := p.backup
	p.mu.Unlock()
	backup.UpdateLogRPC(o)
	p.mu.Lock()
	if len(p.commitLog) < len(o) {
		p.commitLog = o
	}
	p.mu.Unlock()
}

func (b *BackupServer) UpdateLogRPC(log []byte) {
	b.mu.Lock()
	// BUG: this won't work without proposal numbers/configuration numbers
	if !b.isPrimary {
		if len(log) > len(b.opLog) {
			b.opLog = log
			b.commitLog = b.opLog
		}
	}
	b.mu.Unlock()
}
