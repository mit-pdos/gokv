package reconfig

import "github.com/mit-pdos/gokv/grove_ffi"

type AppendArgs struct {
	epoch uint64
	entry LogEntry
	index uint64
}

func EncodeAppendArgs(args *AppendArgs) []byte {
	// TODO:impl
	return nil
}

func DecodeAppendArgs(args []byte) *AppendArgs {
	// TODO:impl
	return nil
}

type Configuration struct {
	replicas []grove_ffi.Address
}

type BecomeReplicaArgs struct {
	epoch      uint64
	startIndex uint64
	log        []LogEntry
}

type BecomePrimaryArgs struct {
	epoch   uint64
	conf    Configuration
	repArgs *BecomeReplicaArgs
}

type GetLogReply struct {
	err        Error
	log        []LogEntry
	startIndex uint64
}
