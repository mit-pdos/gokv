#!/usr/bin/env python3
from os import path
import argparse
import subprocess
import re
import json
import os
import resource
import itertools
import time
import atexit
import signal
from datetime import datetime

from common import *

def num_threads(i):
    if i < 10:
        return 5 + i * 5
    i = i - 10
    return 100 * (i + 1)

    if i < 5:
        return i + 1
    elif i < 25:
        return 5 + (i - 5) * 5
    else:
        return 500 + (i - 25) * 500

def closed_lt(kvname, warmuptime, runtime, valuesize, outfilename, readprop, updateprop, recordcount, thread_fn):
    data = []
    i = 0
    last_good_index = i
    peak_thruput = 0

    while True:
        if i > last_good_index + 5:
            break
        threads = thread_fn(i)

        cleanup_procs()
        start_fresh_single_node_redisraft()
        a = goycsb_bench(kvname, threads, warmuptime, runtime, valuesize, readprop, updateprop, recordcount,
                         ['-p', f"redis.addr={config['serverhost']}:5001"])

        p = {'service': kvname, 'num_threads': threads, 'lts': a}

        data = data + [ p ]
        with open(outfilename, 'a+') as outfile:
            outfile.write(json.dumps(p) + '\n')

        thput = sum([ a[op]['thruput'] for op in a ])

        if thput > peak_thruput:
            last_good_index = i
        if thput > peak_thruput:
            peak_thruput = thput

        # last_thruput = int(thput + 1)
        last_threads = threads

        i = i + 1

    return data

config = {}

def start_fresh_single_node_redisraft():
    os.system("./start-redis.py --ncores 1")

def main():
    atexit.register(cleanup_procs)
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    global config

    config = {
        'read': 0,
        'write': 1.0,
        'keys': 1000,
        'serverhost': '10.10.1.1',
        'warmuptime': 20,
        'runtime': 60,
    }

    outfilepath = global_args.outfile

    closed_lt('rediskv', config['warmuptime'], config['runtime'], 128, outfilepath, config['read'], config['write'], config['keys'], num_threads)

if __name__=='__main__':
    main()
