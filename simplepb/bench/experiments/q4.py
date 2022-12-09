#!/usr/bin/env python3

from os import system as do
import os

print("# finds the peak throughput at each number of cores for redis")
print("ncores, throughput, threads, avglatency")
for ncores in range(1,9):
    do(f"./start-redis.py --ncores {str(ncores)} >/tmp/ephemeral.out 2>/tmp/ephemeral.err")
    peakdata = os.popen("./find-peak.py --onlypeak --benchcmd='./redis-bench-put.py' ").read().rstrip()
    print(f"{ncores}, {peakdata}")
