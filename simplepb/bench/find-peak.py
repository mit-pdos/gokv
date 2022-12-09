#!/usr/bin/env python3

import os
from os import system as do
from parseycsb import *
import sys
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('--onlypeak',
                    help='if this flag is set, then this only prints the peak throughput, not all the intermediate throughputs',
                    action='store_true')
parser.add_argument('--benchcmd',
                    help='The command to run to benchmark a single server (should be either the one for GroveKV or for redis)',
                    default='./bench-put.py')
args = parser.parse_args()

threadcounts = [100] + [200 * (i + 1) for i in range(10)]

runtime = 5
warmuptime = 5
fullwarmuptime = 20
fullruntime = 120

if not args.onlypeak:
    print("threads, throughput, avglatency")

highestThreads = 0
highestThruput = 0
highestThruputLatency = 0

for threads in threadcounts:
    benchoutput = os.popen(f"{args.benchcmd} {threads} --warmup {warmuptime}", 'r', 100)

    ret = ''
    for line in benchoutput:
        if line.find('Takes(s): {0}.'.format(runtime)) != -1:
            ret = line
            benchoutput.close()
            break

    do("killall go-ycsb > /dev/null")
    time, ops, latency = (parse_ycsb_output_totalops(ret))
    thruput = ops/time

    if not args.onlypeak:
        print(f"{threads}, {thruput}, {latency}")

    if thruput > highestThruput:
        highestThruput = thruput
        highestThreads = threads
        highestThruputLatency = latency
    if 1.10 * thruput < highestThruput:
        break # don't bother trying them all


if args.onlypeak:
    print(f"{highestThruput}, {highestThreads}, {highestThruputLatency}")
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
