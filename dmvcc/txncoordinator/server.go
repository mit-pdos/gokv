package txncoordinator

import (
	"fmt"

	"github.com/mit-pdos/gokv/dmvcc/index"
)

type Server struct {
	indexCk *index.Clerk
}

func (s *Server) TryCommit(tid uint64, writes map[uint64]string) bool {
	// FIXME: this whole thing should be a modification of wrbuf

	// acquire locks on everything (i.e. "Prepare")
	var err = uint64(0)

	// This will deadlock.
	for key, _ := range writes {
		err = s.indexCk.AcquireTuple(key, tid)
		if err != 0 {
			break
		}
	}

	if err != 0 {
		fmt.Print("Error\n")
		// FIXME: release locks
		return false
	}

	fmt.Print("Updating\n")
	// now that all "participants" are prepared, transaction can commit
	s.indexCk.UpdateAndRelease(tid, writes)
	return true
}

func MakeServer(indexHost *index.Server) *Server {
	return &Server{
		indexCk: index.MakeClerk(indexHost),
	}
}
