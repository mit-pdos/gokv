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
    returns latency/thruput pairs
    """
    xys = []
    for d in data:
        x = 0.0
        y = 0.0
        for k, v in d['lts'].items():
            # look for reads; if any other operation is found, report an error
            if k.strip() == 'READ':
                x = x + v['thruput']
                y = y + v['avg_latency'] / 1000
            else:
                throw("unimpl")
        xys.append((x, y))
    with open(outfilename, 'w') as f:
        for xy in xys:
                print('{0}, {1}'.format(xy[0], xy[1]), file=f)

os.chdir('/users/upamanyu/gokv/simplepb/bench')
for readpercentage in [1.0, 0.95, 0.5]:
    do('mv /tmp/gokv/grovekv-ro-lts.txt /tmp/grovekv-ro-lts.old')
    do(f'./lt_pb_mult.py -v -e --reads {str(readpercentage)} --outfile /tmp/gokv/grovekv-ro-lts.txt 1>/tmp/pb.out 2>/tmp/pb.err')
    write_lts(read_raw_lt_data("/tmp/gokv/grovekv-ro-lts.txt"), f"./data/readopt/grovekv-{str(int(100*readpercentage))}.dat")

# do('./lt_redis_single.py -v -e --outfile /tmp/gokv/redis-lts.txt 1>/tmp/redis.out 2>/tmp/redis.err')
# write_lts(read_raw_lt_data("/tmp/gokv/redis-lts.txt"), "./data/readopt/redis.dat")
