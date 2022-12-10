#!/usr/bin/env python3

from os import system as do
import os
import json

# Gives redis latency/throughput and GroveKV latency/throughput curves.

os.chdir('/users/upamanyu/gokv/simplepb/bench')
do('mv /tmp/gokv/grovekv-lts.txt /tmp/grovekv-lts.old')
do('mv /tmp/gokv/redis-lts.txt /tmp/redis-lts.old')
do('./lt_pb_single.py -v -e --outfile /tmp/gokv/grovekv-lts.txt 1>/tmp/pb.out 2>/tmp/pb.err')
do('./lt_redis_single.py -v -e --outfile /tmp/gokv/redis-lts.txt 1>/tmp/redis.out 2>/tmp/redis.err')

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
            # look for updates; if any other operation is found, report an error
            if k == 'UPDATE':
                x = x + v['thruput']
                y = y + v['avg_latency'] / 1000
                xys.append((x,y))
            else:
                throw("unimpl")
    with open(outfilename, 'w') as f:
        for xy in xys:
                print('{0}, {1}'.format(xy[0], xy[1]), file=f)

write_lts(read_raw_lt_data("/tmp/gokv/grovekv-lts.txt"), "./data/redis_vs_grove/grovekv.dat")
write_lts(read_raw_lt_data("/tmp/gokv/redis-lts.txt"), "./data/redis_vs_grove/redis.dat")
