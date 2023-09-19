#!/usr/bin/env python3

print("""
# Measures the peak throughput of GroveKV with an increasing number of servers,
# 1-3. Varies read ratio is between 0, 0.5, 0.95, and 1.0.
# Data is put in `./data/multi`.
""")

import os
from os import system as do
from bench.find_peak import find_peak

os.chdir(os.path.expanduser('~/gokv/simplepb/bench'))

threadcounts = [50, 100, 150, 200, 250, 300, 400, 500, 600, 800, 1000, 1200]

# TODO: add a cache for this test
do("mkdir -p ./data/multi")
for reads in [0.0, 0.5, 0.95, 1.0]:
    with open(f'./data/multi/servers{int(reads*100)}.dat', 'w') as f:
        pass
    for nreplicas in range(1,4):
        # find peak throughput at this configuration
        highestThruput, highestThreads = find_peak(8, nreplicas, reads, threadcounts, outfilename="/tmp/peaks.txt")
        with open(f'./data/multi/servers{int(reads*100)}.dat', 'a+') as f:
            print(f"{nreplicas}, {highestThruput} % at threads={highestThreads}", file=f)
