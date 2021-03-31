Must use `ulimit -n 100000` or some such to allow enough redis clients to be
usable. Redis will limit its max number of clients based on how many file
descriptors it is allowed open.
Want to have enough clients for redis benchmark.

go-ycsb's interface is not quite what we want.
It expects each key to have a corresponding row, and each row has some
(field,value) pairs. We only care about a KV map.

We want an interface that simply does puts and gets.

Want to measure scale-up on the sharded KV store; if we add new machines, does
perf increase as expected?

Could even think about elastic scale-up.

To get KV benchmark, need to only run YCSB with 1 field, and only use the value,
while ignoring the name of the field:
https://github.com/secure-foundations/veribetrkv-osdi2020/blob/053dd16ab01939843a213e0590983fd22e2d0d4d/docker-ssd/src/veribetrkv-linear/ycsb/YcsbMain.cpp#L161
Perhaps a bit hacky. Can re-write later if worthwhile.

Q: Have observed 
  Max(us): 90834, 99th(us): 47000, 99.9th(us): 78000, 99.99th(us): 91000
on redis
A: because the max is rounded to the nearest millisecond (buckets of latencies).
