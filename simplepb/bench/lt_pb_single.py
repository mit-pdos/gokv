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
from .goycsb import *

def closed_lt(kvname, config, thread_fn):
    i = 0
    last_good_index = i
    peak_thruput = 0

    while True:
        if i > last_good_index + 5:
            break
        threads = thread_fn(i)

        # restart and reload every single time
        start_single_core_single_node_kv_system()
        goycsb_load(kvname, 10, config['valuesize'], config['recordcount'],
                    ['-p', f"pbkv.configAddr={config['serverhost']}:12000"])

        a = goycsb_bench(kvname, threads, config['warmuptime'], config['runtime'],
                         config['valuesize'], config['reads'], config['writes'],
                         config['recordcount'],
                         ['-p', f"pbkv.configAddr={config['serverhost']}:12000"])
        p = {'service': kvname, 'num_threads': threads, 'lts': a}

        with open(config['outfilename'], 'a+') as outfile:
            outfile.write(json.dumps(p) + '\n')

        thput = sum([ a[op]['thruput'] for op in a ])

        if thput > peak_thruput:
            last_good_index = i
        if thput > peak_thruput:
            peak_thruput = thput

        last_threads = threads

        i = i + 1

    return

def start_single_core_single_node_kv_system():
    os.system("./start-pb.py --ncores 1 1")

def run(outfilepath, readratio, threads_fn, warmuptime=30, runtime=120):
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    config = {
        'outfilename': outfilepath,
        'reads': readratio,
        'writes': 1 - readratio,
        'recordcount': 1000,
        'serverhost': '10.10.1.4',
        'warmuptime': warmuptime,
        'runtime': runtime,
        'valuesize': 128,
    }

    closed_lt('pbkv', config, threads_fn)
    cleanup_procs()
