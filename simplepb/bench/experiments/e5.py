#!/usr/bin/env python3

from os import system as do
import os
import time
from parseycsb import *

gobin='/usr/local/go/bin/go'

def collect_samples_one_pb(reads, num_samples, nthreads):
    samples = []
    for i in range(num_samples):
        do(f"./start-pb.py --ncores 1 1 > /tmp/ephemeral.out 2>/tmp/ephemeral.err")

        do(f"""~/go-ycsb/go-ycsb load pbkv -P /users/upamanyu/gokv/simplepb/bench/pbkv_workload --threads 200 --target -1 \
        --interval 1000 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion={str(reads)} \
        -p updateproportion={str(1.0 - reads)} -p warmuptime=30 -p recordcount=1000 -p pbkv.configAddr=10.10.1.4:12000 > /tmp/run.txt
        """)

        o = os.popen(f"""~/go-ycsb/go-ycsb run pbkv -P /users/upamanyu/gokv/simplepb/bench/pbkv_workload --threads {nthreads} --target -1 \
        --interval 1000 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion={str(reads)} \
        -p updateproportion={str(1.0 - reads)} -p warmuptime=30 -p recordcount=1000 -p pbkv.configAddr=10.10.1.4:12000 > /tmp/run.txt
        """)

        time.sleep(90) # let it run for 60 seconds post-warmup
        do("killall go-ycsb")
        data = ''
        with open("/tmp/run.txt") as f:
            data = f.read()

        thput = 0
        gotfirst = False
        for line in reversed(data.splitlines()):
            t,ops,latency, = (parse_ycsb_output_totalops(line + "\n"))
            if t is None:
                if gotfirst:
                    break
                else:
                    continue
            gotfirst = true
            thput += float(ops)/t
        samples = samples + [thput]

    return samples

def collect_samples_one_redis(reads, num_samples, nthreads):
    samples = []
    for i in range(num_samples):
        do(f"./start-redis.py --ncores 1 > /tmp/ephemeral.out 2>/tmp/ephemeral.err")

        do(f""" ~/go-ycsb/go-ycsb load rediskv -P /users/upamanyu/gokv/simplepb/bench/rediskv_workload \
        --threads 200 --target -1 --interval 1000 -p operationcount=4294967295 -p fieldlength=128 \
        -p requestdistribution=uniform -p readproportion=0.0 -p updateproportion=1.0 \
        -p warmuptime=30 -p recordcount=1000 -p redis.addr=10.10.1.1:5001 > /tmp/run.txt
        """)

        o = os.popen(f""" ~/go-ycsb/go-ycsb run rediskv -P /users/upamanyu/gokv/simplepb/bench/rediskv_workload \
        --threads {nthreads} --target -1 --interval 1000 -p operationcount=4294967295 -p fieldlength=128 \
        -p requestdistribution=uniform -p readproportion={str(reads)} -p updateproportion={str(1.0 - reads)} \
        -p warmuptime=30 -p recordcount=1000 -p redis.addr=10.10.1.1:5001 > /tmp/run.txt
        """)

        time.sleep(90) # let it run for 60 seconds
        do("killall go-ycsb")
        data = ''
        with open("/tmp/run.txt") as f:
            data = f.read()

        thput = 0
        gotfirst = False
        for line in reversed(data.splitlines()):
            t,ops,latency, = (parse_ycsb_output_totalops(line + "\n"))
            if t is None:
                if gotfirst:
                    break
                else:
                    continue
            gotfirst = true
            thput += float(ops)/t
        samples = samples + [thput]

    return samples

s = collect_samples_one_pb(0, 20, 1500)
print('pb, 0 reads 10 samples')
print(s)

s = collect_samples_one_redis(0, 20, 1500)
print('redis, 0 reads 10 samples')
print(s)

s = collect_samples_one_pb(0.5, 20, 1100)
print('pb, 50 reads 10 samples')
print(s)

s = collect_samples_one_redis(0.5, 20, 1100)
print('redis, 50 reads 10 samples')
print(s)

s = collect_samples_one_pb(0.95, 20, 50)
print('pb, 95 reads 10 samples, 50 threads')
print(s)

s = collect_samples_one_redis(0.95, 20, 1200)
print('redis, 95 reads 10 samples, 1200 threads')
print(s)

s = collect_samples_one_pb(0.95, 20, 600)
print('pb, 95 reads 10 samples, 600 threads')
print(s)
