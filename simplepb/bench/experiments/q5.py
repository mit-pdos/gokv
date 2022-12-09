#!/usr/bin/env python3

from os import system as do
import os

print("""# Prints out the instantaneous latency and throughput of GroveKV with a
         # crash and then a reconfiguration in the middle""")

print("ncores, ")
for ncores in range(1,9):
    do(f"./start-pb.py 1 --ncores 8> /tmp/ephemeral.out 2>/tmp/ephemeral.err")
    print(f"{ncores}, {peakdata}")

    peakdata = os.popen("./bench-put.py 2>/tmp/peak.err").read().rstrip()
    do(f"reconfigure")
