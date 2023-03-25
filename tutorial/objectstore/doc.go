// A distributed object store.
//
// The high-level abstraction this system provides is a mapping from small,
// string keys to large byte values. A single value can be efficiently streamed
// for reads and writes. Modifications must overwrite the entire key.
//
// The state of the system is spread out over chunk servers that store chunks
// (imagine 4KB or perhaps 64KB in size) of data by content hash and a directory
// service that tracks which chunks are part of each key's value and which chunk
// server holds that chunk. In-progress writes are also tracked by the directory
// service until they are committed to a value, at which point they become
// visible to future client reads.
//
// The implementation is divided into three pieces: a _directory service_ (which
// is a single, dedicated server), _chunk servers_, and a _client library_ that
// coordinates.
//
// The implementation is designed to handle message duplication but not
// failures, and in particular a failure or loss of data at the directory
// service would lose track of essentially all data in the system, since the
// organization of chunk data would be completely lost. Furthermore chunks are
// not currently replicated.
//
// The directory service can, in principle, attempt to balance data among chunk
// servers. Currently no intelligent placement policy is implemented;
// furthermore some additional Transfer operations are needed to allow the
// directory server to dynamically rebalance among chunk servers. However, the
// design is agnostic to where a chunk is stored; the directory service
// exclusively tracks this information and informs clients where to find data.
//
// The concept of a ChunkHandle is useful to explain the data stored and RPCs. A
// ChunkHandle is a combination of a content hash and server. It is enough
// information to lookup a chunk worth of data, and since chunks are immutable
// is effectively a stand-in for that data.
//
// The flow for a write is the following:
//   - Client sends a PrepareWrite RPC to the dir service. This allocates a fresh
//     WriteID to hold the unpublished data, and gives it some chunk servers to
//     store the data at (this is potentially where balancing decisions come in).
//   - The client can now stream writes for this value, sending a sequence of
//     WriteChunk RPC to each chunk server. These can be issued asynchronously
//     (they could be performed out-of-order at the chunk servers but the client
//     library does not support this).
//   - In response to a Writechunk, the chunk servers store the chunk locally,
//     then send a RecordChunk RPC to the directory service that incorporates this
//     chunk (as a ChunkHandle) into the ongoing WriteID at the correct
//     index. This is performed synchronously so when the WriteChunk returns to
//     the client the chunk is recorded at the directory service, and in
//     particular when all WriteChunk RPCs complete the prepared WriteID has the
//     data the client intended to write to this key.
//   - Finally the client sends a FinishWrite RPC to the dir service, which puts
//     the ChunkHandles in that ongoing write into the mapping for the key. When
//     this is complete the write is logically complete.
//
// The flow for a read is the following:
//   - Client sends a PrepareRead RPC to the dir service. This returns a list of
//     ChunkHandles, in the order they appear in that key's value. Note that this
//     does not involve WriteIDs (ongoing writes) at all.
//   - The client sends GetChunk RPCs to the chunk servers for each ChunkHandle.
//     These could be issued in parallel and assembled in the correct order, but the
//     library currently performs them synchronously.
//
// We have not yet considered deleting ongoing writes after they are complete.
// This allows them to be updated by duplicate RecordChunks, which is benign but
// useless.
package objectstore
