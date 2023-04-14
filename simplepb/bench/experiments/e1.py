#!/usr/bin/env python3

from os import system as do
import os
import json

# Gives redis latency/throughput and GroveKV latency/throughput curves.

def read_raw_lt_data(infilename):
    with open(infilename, 'r') as f:
        data = []
        for line in f:
            data.append(json.loads(line))
        return data
    return None

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
            if k == 'UPDATE':
                wthru = v['thruput']
                wlat = v['avg_latency'] / 1000
            elif k == 'READ':
                rthru = v['thruput']
                rlat = v['avg_latency'] / 1000
            else:
                throw("unimpl")
        xys.append((d['num_threads'], rlat, rthru, wlat, wthru))
    with open(outfilename, 'w') as f:
        for xy in xys:
                print(f"{xy[0]}, {xy[1]}, {xy[2]}, {xy[3]}, {xy[4]}", file=f)

os.chdir('/users/upamanyu/gokv/simplepb/bench')
do('mv /tmp/gokv/grovekv-lts.txt /tmp/grovekv-lts.old')
do('mv /tmp/gokv/redis-lts.txt /tmp/redis-lts.old')

for readratio in [0.0, 0.5, 0.95]:
    id = str(int(100 * readratio))
    do(f'./lt_pb_single.py -v -e --reads {str(readratio)} --outfile /tmp/gokv/grovekv-lts.txt 1>/tmp/pb.out 2>/tmp/pb.err')
    do(f'./lt_redis_single.py -v -e --reads {str(readratio)} --outfile /tmp/gokv/redis-lts.txt 1>/tmp/redis.out 2>/tmp/redis.err')
    do(f"cp /tmp/gokv/grovekv-lts.txt ./data/redis_vs_grove/grovekv-lts-{id}.txt")
    do(f"cp /tmp/gokv/redis-lts.txt ./data/redis_vs_grove/redis-lts-{id}.txt")
    write_lts(read_raw_lt_data("/tmp/gokv/grovekv-lts.txt"), f"./data/redis_vs_grove/grovekv-{id}.dat")
    write_lts(read_raw_lt_data("/tmp/gokv/redis-lts.txt"), f"./data/redis_vs_grove/redis-{id}.dat")
