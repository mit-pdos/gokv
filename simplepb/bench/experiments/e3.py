#!/usr/bin/env python3

# Run between 1 and 3 servers.
# Measures the peak throughput of each.
# Read ratio is 0.5, write ratio is 0.5.

import os
from os import system as do

os.chdir('/users/upamanyu/gokv/simplepb/bench')

for ncores in range(1,9):
    # find peak throughput at this configuration
    with open("./data/peak.txt", "a+") as f:
        f.write(f"# Running with {ncores} cores on each server\n")

    do(f"./find-peak.py --ncores {ncores} --nreplicas 3 --reads 0.5 2>/tmp/peak.err >> ./data/peak.txt")
