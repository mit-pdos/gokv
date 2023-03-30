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

def closed_lt(kvname, warmuptime, runtime, valuesize, outfilename, readprop, updateprop, recordcount, thread_fn):
    i = 0
    last_good_index = i
    peak_thruput = 0

    while True:
        if i > last_good_index + 7:
            break
        threads = thread_fn(i)

        # restart every single time
        start_kv_system()

        # start_single_core_single_node_kv_system()
        a = goycsb_bench(kvname, threads, warmuptime, runtime, valuesize, readprop, updateprop, recordcount,
                         ['-p', f"pbkv.configAddr={config['serverhost']}:12000"])
        p = {'service': kvname, 'num_threads': threads, 'lts': a}

        with open(outfilename, 'a+') as outfile:
            outfile.write(json.dumps(p) + '\n')

        thput = sum([ a[op]['thruput'] for op in a ])

        if thput > peak_thruput:
            last_good_index = i
        if thput > peak_thruput:
            peak_thruput = thput

        last_threads = threads

        i = i + 1

    return

def start_kv_system():
    os.system("./start-pb.py --ncores 8 3")

config = {}

def main():
    atexit.register(cleanup_procs)
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    global config

    readratio = float(global_args.reads)
    config = {
        'read': readratio,
        'write': 1 - readratio,
        'keys': 1000,
        'serverhost': '10.10.1.4',
        'warmuptime': 10,
        'runtime': 10,
    }

    outfilepath = global_args.outfile
    closed_lt('pbkv', config['warmuptime'], config['runtime'], 128, outfilepath, config['read'], config['write'], config['keys'], num_threads)
    cleanup_procs()

if __name__=='__main__':
    main()
