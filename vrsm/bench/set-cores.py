#!/usr/bin/env python3

from os import system as do
import sys

cpus = [0,2,4,6,8,10,12,14]
numCores = int(sys.argv[1])
if numCores == 0:
    print("Need at least 1 core on")
    sys.exit(1)

for i in cpus[1:numCores]:
    do(f"echo 1 | sudo tee /sys/devices/system/cpu/cpu{i}/online")

for i in cpus[numCores:]:
    do(f"echo 0 | sudo tee /sys/devices/system/cpu/cpu{i}/online")
