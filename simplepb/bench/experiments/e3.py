#!/usr/bin/env python3

# Run between 1 and 3 servers, varying number of cores on each server.
# Measures the peak throughput of each with increasing number of cores.
# Read ratio is 0.95, write ratio is 0.05.

import os
from os import system as do
from bench.find_peak import find_peak

os.chdir(os.path.expanduser('~/gokv/simplepb/bench'))

threadcounts = [50, 100, 150, 200, 250, 300, 400, 500, 600, 800, 1000, 1200]

for reads in [0.0, 0.5, 0.95, 1.0]:
    for nreplicas in range(1,4):
        # find peak throughput at this configuration
        highestThruput, highestThreads = find_peak(8, nreplicas, reads, threadcounts, outfilename="/tmp/peaks.txt")
        with open(f'./data/multi/servers{int(reads*100)}.dat', 'a+') as f:
            print(f"{nreplicas}, {highestThruput} % at threads={highestThreads}", file=f)
