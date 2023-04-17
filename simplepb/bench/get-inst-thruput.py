#!/usr/bin/env python3

from parseycsb import *
import sys

data = ''
with open(sys.argv[1]) as f:
    data = f.read()

# averaged over the last 1second, so keep 5 last data points
lastops = [0 for i in range(6)]
for line in data.splitlines():
    time,ops,latency, = (parse_ycsb_output_totalops(line + "\n"))
    if time is None:
        continue
    lastops = lastops[1:] + [ops]
    print(f"{time}, {lastops[-1] - lastops[0]}")
