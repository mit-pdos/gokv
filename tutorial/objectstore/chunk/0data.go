package chunk

type WriteChunkArgs struct {
	WriteId WriteID
	Chunk   []byte
	Index   uint64
}

func MarshalWriteChunkArgs(args WriteChunkArgs) []byte {
	panic("TODO: marshalling")
}

func ParseWriteChunkArgs(data []byte) WriteChunkArgs {
	panic("TODO: marshalling")
}
