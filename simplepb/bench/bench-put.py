#!/usr/bin/env python3
import os
from os import system as do
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('threads', metavar='nthread', type=int,
                    help='number of threads')
parser.add_argument('--warmup', type=int, default=10)
parser.add_argument('--interval', type=int, default=100)
args = parser.parse_args()

gobin='/usr/local/go/bin/go'
os.chdir('/users/upamanyu/go-ycsb/')
do(f"""{gobin} run ./cmd/go-ycsb run pbkv
    -P /users/upamanyu/gokv/simplepb/bench/pbkv_workload
    --threads {str(args.threads)}
    --target -1
    --interval {str(args.interval)}
    -p operationcount={str(2**32 - 1)}
    -p fieldlength=128
    -p requestdistribution=uniform
    -p readproportion=0.0
    -p updateproportion=1.0
    -p warmuptime={args.warmup}
    -p recordcount=1000
    -p pbkv.configAddr=10.10.1.4:12000
""".replace('\n', ' '))
