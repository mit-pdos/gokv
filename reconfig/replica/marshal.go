package replica

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
	Replicas []grove_ffi.Address
}

func EncodeConfiguration(conf *Configuration) []byte {
	panic("impl")
}

func DecodeConfiguration(conf_enc []byte) *Configuration {
	panic("impl")
}

type BecomeReplicaArgs struct {
	Epoch      uint64
	StartIndex uint64
	Log        []LogEntry
}

type BecomePrimaryArgs struct {
	Epoch uint64
	Conf  Configuration
}

type GetLogReply struct {
	err        Error
	log        []LogEntry
	startIndex uint64
}
