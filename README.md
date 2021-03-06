# Starting server
`go run ./cmd/srv` will start the GoKV server. It will put durable state in a
file called `kvdur` in the root project directory.

After that, you can run `go run ./bench` to run benchmarks against the GoKV
server.

To run benchmarks against redis, start redis on port 6379 before running bench.
