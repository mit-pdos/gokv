#!/usr/bin/env python3

# Run 3 servers, varying number of cores on each server.
# Measures the peak throughput of each with increasing number of cores.
# Read ratio is 0.95, write ratio is 0.05.

import os
from os import system as do

os.chdir('/users/upamanyu/gokv/simplepb/bench')

for ncores in range(1,9):
    # find peak throughput at this configuration
    with open("./data/multicore.txt", "a+") as f:
        f.write(f"# Running with {ncores} cores on each server\n")

    do(f"./find-peak.py --ncores {ncores} --nreplicas 3 --reads 0.95 2>/tmp/multicore.err >> ./data/multicore.txt")
