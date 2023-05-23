#!/usr/bin/env python3

import csv

def get_highest(fname):
    with open(fname, 'r') as f:
        reader = csv.reader(f)
        highest = 0
        for row in reader:
            thput = float(row[2]) + float(row[4])
            highest = max(thput, highest)
        return highest
    return "No file found"

for readratio in [0, 50, 95]:
    print(f"redis {readratio} -> ", get_highest(f"redis-{readratio}.dat"))
    print(f"grovekv {readratio} -> ", get_highest(f"grovekv-{readratio}.dat"))
