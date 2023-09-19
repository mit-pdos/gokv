#!/usr/bin/env python3

# Run between 1 and 3 servers.
# Measures the peak throughput of each.
# Read ratio varies also.

import os
import time
from os import system as do

os.chdir('/users/upamanyu/gokv/simplepb/bench')

starttime = int(time.time())
filename = f"./data/peak-{str(starttime)}.txt"
for reads in [0.0, 0.5, 0.95, 1.0]:
    for nreplicas in range(1,4):
        # find peak throughput at this configuration
        with open(filename, "a+") as f:
                f.write(f"# Running with {nreplicas} servers, {reads} reads\n")

        do(f"./find-peak.py --ncores 8 --nreplicas {nreplicas} --reads {str(reads)} 2>/tmp/peak.err >> {filename}")
