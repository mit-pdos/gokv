#!/usr/bin/env -S python3 -u

import os
from os import system as do
from parseycsb import *
import sys
import argparse
import json
from bench_ops_multiclient import *

parser = argparse.ArgumentParser()
parser.add_argument('--onlypeak',
                    help='if this flag is set, then this only prints the peak throughput, not all the intermediate throughputs',
                    action='store_true')
parser.add_argument("--reads",
                    help="percentage of ops that are reads (between 0.0 and 1.0)",
                    required=False,
                    default=0.0)
parser.add_argument("--ncores", help="number of cores per server", required=True)
parser.add_argument("--nreplicas", help="number of servers", required=True)
parser.add_argument('--benchcmd',
                    help='The command to run to benchmark a single server (should be either the one for GroveKV or for redis)',
                    default='./bench-ops-multiclient.py')
args = parser.parse_args()

threadcounts = [50, 100, 150, 200, 250] + [200 * (i + 2) for i in range(3)] + [500 * (i + 3) for i in range (15)]
client_machines = [5, 6, 7]

runtime = 20
warmuptime = 20
cooldown = 10
fullwarmuptime = 20
fullruntime = 120
recordcount = 1000

if not args.onlypeak:
    print("threads, write throughput, write avglatency, read throughput, read avglatency")

highestThreads = 0
highestThruput = 0

def restart_system():
    do_quiet("./stop-pb.py")
    do_quiet(f"./start-pb.py --ncores {args.ncores} {args.nreplicas}")
    do_quiet(f"./bench-load.py --recordcount {recordcount}")
    return

for threads in threadcounts:
    restart_system()
    ms = get_thruput(client_machines, threads, args.reads, recordcount, runtime, cooldown, warmuptime)

    wthruput = 0.0
    wlatency = 0.0
    rthruput = 0.0
    rlatency = 0.0
    for m in ms:
        if 'UPDATE' in m["lts"]:
            wthruput += m["lts"]['UPDATE']["thruput"]
            wlatency += m["lts"]['UPDATE']["avg_latency"]
        if 'READ' in m["lts"]:
            rthruput += m["lts"]['READ']["thruput"]
            rlatency += m["lts"]['READ']["avg_latency"]

    rlatency = rlatency/len(ms)
    wlatency = wlatency/len(ms)

    if not args.onlypeak:
        print(f"{threads}, {wthruput}, {wlatency}, {rthruput}, {rlatency}")

    thruput = rthruput + wthruput

    if thruput > highestThruput:
        highestThruput = thruput
        highestThreads = threads
    if 1.10 * thruput < highestThruput:
        break # don't bother trying them all

if args.onlypeak:
    print(f"{highestThruput}, {highestThreads}")
else:
    print(f"# Highest throughput {highestThruput} at {highestThreads} threads")

sys.exit(0) # Don't bother with running for longer for now

# Run the nthread that gave the highest throughput for longer
benchoutput = os.popen(f"{args.benchcmd} {highestThreads} --warmup {fullwarmuptime}", 'r', 100)
ret = ''
for line in benchoutput:
    if line.find('Takes(s): {0}.'.format(fullruntime)) != -1:
        ret = line
        benchoutput.close()
        break
time, ops, latency = (parse_ycsb_output_totalops(ret))
thruput = ops/time

print(f"# sustained throughput {thruput} at {highestThreads} threads")
