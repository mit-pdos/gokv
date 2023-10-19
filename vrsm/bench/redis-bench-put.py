#!/usr/bin/env python3
import os
from os import system as do
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('threads', metavar='nthread', type=int,
                    help='number of threads')
parser.add_argument('--warmup', type=int, default=10)
args = parser.parse_args()

gobin='/usr/local/go/bin/go'
os.chdir('/users/upamanyu/go-ycsb/')
do(f"""{gobin} run ./cmd/go-ycsb run rediskv
    -P /users/upamanyu/gokv/vrsm/bench/rediskv_workload
    --threads {str(args.threads)}
    --target -1
    --interval 100
    -p operationcount={str(2**32 - 1)}
    -p fieldlength=128
    -p requestdistribution=uniform
    -p readproportion=0.0
    -p updateproportion=1.0
    -p warmuptime={args.warmup}
    -p recordcount=1000
    -p redis.addr=10.10.1.1:5001
""".replace('\n', ' '))
