#!/usr/bin/env python3

import os
from os import system as do
from parseycsb import *
import sys

threadcounts = [10, 50, 100, 200, 500, 1000, 1500, 2000, 3000]

runtime = 10
warmuptime = 20
fullwarmuptime = 20
fullruntime = 120

print("threads, throughput, avglatency")

highestThreads = 0
highestThruput = 0

for threads in threadcounts:
    benchoutput = os.popen(f"./bench-put.py {threads} --warmup {warmuptime}", 'r', 100)

    ret = ''
    for line in benchoutput:
        if line.find('Takes(s): {0}.'.format(runtime)) != -1:
            ret = line
            benchoutput.close()
            break

    do("killall go-ycsb > /dev/null")
    time, ops, latency = (parse_ycsb_output_totalops(ret))
    thruput = ops/time
    if thruput > highestThruput:
        highestThruput = thruput
        highestThreads = threads

    print(f"{threads}, {thruput}, {latency}")

print(f"# highest thruput {highestThruput} at {highestThreads} threads")

sys.exit(0) # Don't bother with running for longer for now

# Run the nthread that gave the highest throughput for longer
benchoutput = os.popen(f"./bench-put.py {highestThreads} --warmup {fullwarmuptime}", 'r', 100)
ret = ''
for line in benchoutput:
    if line.find('Takes(s): {0}.'.format(fullruntime)) != -1:
        ret = line
        benchoutput.close()
        break
time, ops, latency = (parse_ycsb_output_totalops(ret))
thruput = ops/time

print(f"# sustained throughput {thruput} at {highestThreads} threads")
