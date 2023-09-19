#!/usr/bin/env python3

from os import system as do
import os

print("# finds the peak throughput of a single GroveKV server with varying number of cores")
print("ncores, ")
for ncores in range(1,9):
    do(f"./start-pb.py 1 --ncores {str(ncores)} > /tmp/ephemeral.out 2>/tmp/ephemeral.err")
    peakdata = os.popen("./find-peak.py --onlypeak 2>/tmp/peak.err").read().rstrip()
    print(f"{ncores}, {peakdata}")
