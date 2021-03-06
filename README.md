`go run ./cmd/srv` will start the GoKV server. It will put durable state in a
file called `kvdur` in the root project directory.

After that, you can run `go run ./bench` to run benchmarks against the GoKV
server.

To start redis, run `/path/to/redis/src/redis-server ./bench/redis.conf` to run
redis with AOF persistence; its snapshot at aof files get saved in
`./redis-db`.

Then, you can run the redis put benchmark in `./bench`
