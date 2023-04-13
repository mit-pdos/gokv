#!/usr/bin/env python3
import os
from os import system as do
import argparse
import json
import time

parser = argparse.ArgumentParser()
parser.add_argument('threads', metavar='nthread', type=int,
                    help='number of threads')
parser.add_argument('--warmup', type=int, default=10)
parser.add_argument('--cooldown', type=int, default=10)
parser.add_argument("--reads",
                    help="percentage of ops that are reads (between 0.0 and 1.0)",
                    required=False,
                    default=0.0)
parser.add_argument('--recordcount', type=int, default=1000)
args = parser.parse_args()

other_client_machines = [3]

benchcmd = f"./bench-ops.py {args.threads} --reads {args.reads} 1>/tmp/bench.out 2>/tmp/bench.err"
# --recordcount {args.recordcount} --warmup {args.warmup} --cooldown {args.cooldown}

# start bench-ops.py on other machines
for m in other_client_machines:
    do(f"""ssh upamanyu@node{m} <<ENDSSH
    cd /users/upamanyu/gokv/simplepb/bench/;
    nohup {benchcmd} &
ENDSSH
    """)

# start local bench-ops.py; relying on the other one still warming up for this
# to not mess up the numbers
do(benchcmd)

bench_data = []
with open("/tmp/bench.out", "r") as f:
    bench_data += [json.loads(f.read())]

time.sleep(5)
# now, copy the output file from the other machine
for m in other_client_machines:
    do(f"scp upamanyu@node{m}:/tmp/bench.out /tmp/bench{m}.out")
    with open(f"/tmp/bench{m}.out", "r") as f:
        bench_data += [json.loads(f.read())]

print(bench_data)
