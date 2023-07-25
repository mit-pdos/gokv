#!/usr/bin/env -S python3

# Generates raw data for latency/throughput curves for Redis and GroveKV for
# various read ratios.
# Puts final data in `gokv/simplepb/bench/data/redis_vs_grove/`.
# This is intended to help find the peak throughput of the two systems and the
# latency under low load.

from os import system as do
import os
import json
import sys
from bench import lt_pb_single

def write_lts(data, outfilename):
    """
    Assumes data is in format
    [ (kvname, numthreads, { 'OPERATION_TYPE': (throughput in ops/sec, latency in us), ... } ),  ... ]
    returns (numthreads, read_latency, read_thruput, write_latency, write_thruput)
    """
    xys = []
    for d in data:
        wlat = 0.0
        wthru = 0.0

        rlat = 0.0
        rthru = 0.0

        for k, v in d['lts'].items():
            # look for updates; if any other operation is found, report an error
            if k == 'TOTAL':
                wthru = v['thruput']
                wlat = v['avg_latency'] / 1000
            # elif k == 'READ':
                # rthru = v['thruput']
                # rlat = v['avg_latency'] / 1000
            else:
                throw("unimpl")
        xys.append((d['num_threads'], rlat, rthru, wlat, wthru))
    with open(outfilename, 'w') as f:
        for xy in xys:
                print(f"{xy[0]}, {xy[1]}, {xy[2]}, {xy[3]}, {xy[4]}", file=f)

os.chdir(os.path.expanduser('~/gokv/simplepb/bench'))

def num_threads(i):
    if i < 1:
        return 1
    i = i - 1
    if i < 10:
        return 5 + i * 5
    i = i - 10
    return 100 * (i + 1)

for readratio in [0.0, 0.5, 0.95]:
    do('mkdir /tmp/gokv')
    do('mv /tmp/gokv/grovekv-lts.txt /tmp/grovekv-lts.old')
    do('mv /tmp/gokv/redis-lts.txt /tmp/redis-lts.old')
    id = str(int(100 * readratio))
    # do(f'./lt_pb_single.py -v -e --reads {str(readratio)} --outfile /tmp/gokv/grovekv-lts.txt 1>/tmp/pb.out 2>/tmp/pb.err')
    # do(f'./lt_redis_single.py -v -e --reads {str(readratio)} --outfile /tmp/gokv/redis-lts.txt 1>/tmp/redis.out 2>/tmp/redis.err')
    grove_data = lt_pb_single.run("/tmp/gokv/grovekv-lts.txt", readratio, num_threads,
                                  warmuptime=10, runtime=20)
    redis_data = lt_pb_single.run("/tmp/gokv/redis-lts.txt", readratio, num_threads,
                                  warmuptime=10, runtime=20)

    do(f"cp /tmp/gokv/grovekv-lts.txt ./data/redis_vs_grove/grovekv-lts-{id}.txt")
    do(f"cp /tmp/gokv/redis-lts.txt ./data/redis_vs_grove/redis-lts-{id}.txt")
    write_lts(grove_data, f"./data/redis_vs_grove/grovekv-{id}.dat")
    write_lts(redis_data, f"./data/redis_vs_grove/redis-{id}.dat")
