#!/usr/bin/env python3

from parseycsb import *
import sys

data = ''
with open(sys.argv[1]) as f:
    data = f.read()

# averaged over the last 1second, so keep 5 last data points
lastops = [0 for i in range(6)]
n = 0
# XXX: the last line is
for line in data.splitlines():
    if line.startswith("Run finished"):
        break
    time,ops,latency, = (parse_ycsb_output_totalops(line + "\n"))
    if time is None:
        continue
    lastops = lastops[1:] + [ops]
    n += 1
    if n > 5: # skip the first few running averages that include negative time.
        print(f"{time}, {lastops[-1] - lastops[0]}")
