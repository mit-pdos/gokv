## Running
`go run ./cmd/srv` will start the GoKV server. It will put durable state in a
file called `kvdur` in the root project directory.

After that, you can run `go run ./bench` to run benchmarks against the GoKV
server.

To start redis, run `/path/to/redis/src/redis-server ./bench/redis.conf` to run
redis with AOF persistence; its snapshot at aof files get saved in
`./redis-db`.

Then, you can run the redis put benchmark in `./bench`

## Goosing
An annoying thing about goosing: need to put lockservice/ in `$GOPATH/src/...`.
I just did that by temporarily getting rid of `go.mod` and doing `go get`.

With go.mod, it'll end up in `$GOPATH/pkg/mod/...`, which makes it so that
goose has to be run from the directory with the `go.mod` file, instead of from
the perennial directory.

Run `export GOPRIVATE=github.com/mit-pdos` before `go get`.
