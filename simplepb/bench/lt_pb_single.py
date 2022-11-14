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

from common import *

def num_threads(i):
    if i < 5:
        return i + 1
    elif i < 25:
        return 5 + (i - 5) * 5
    else:
        return 500 + (i - 25) * 500

def closed_lt(kvname, valuesize, outfilename, readprop, updateprop, recordcount, thread_fn, benchcpus):
    data = []
    i = 0
    last_good_index = i
    peak_thruput = 0
    # last_thruput = 10000
    # last_threads = 10

    while True:
        if i > last_good_index + 5:
            break
        threads = thread_fn(i)

        a = goycsb_bench(kvname, threads, 20, valuesize, readprop, updateprop, recordcount, benchcpus)
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

def start_config_server():
    # FIXME: core pinning
    start_command(many_cpus(["go", "run", "./cmd/config", "-port", "12000"], config['configcpus']), cwd=simplepbdir)

def start_one_kv_server():
    # FIXME: core pinning
    # delete kvserver.data file
    run_command(["rm", "durable/single_kvserver.data"], cwd=simplepbdir)
    start_command(many_cpus(["go", "run", "./cmd/kvsrv", "-filename", "single_kvserver.data", "-port", "12100"], config['kvcpus']), cwd=simplepbdir)

def start_single_node_kv_system():
    start_config_server()
    start_one_kv_server()
    time.sleep(1.0)
    # tell the config server about the initial config
    start_command(["go", "run", "./cmd/admin", "-conf", "0.0.0.0:12000",
                   "init", "0.0.0.0:12100"], cwd=simplepbdir)

config = {}

def main():
    atexit.register(cleanup_procs)
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    global config

    config = {
        'read': 0,
        'write': 1.0,
        'keys': 1000,
        'clientcpus': '4-7',
        'configcpus': '0',
        'kvcpus': '1',
    }

    # time.sleep(1000000)
    start_single_node_kv_system()
    closed_lt('pbkv', 128, path.join(global_args.outdir, 'pb-kvs.jsons'), config['read'], config['write'], config['keys'], num_threads, config['clientcpus'])
    # closed_lt('rediskv', 128, path.join(global_args.outdir, 'redis-kvs.jsons'), config['read'], config['write'], config['keys'], num_threads, config['clientcores'])

if __name__=='__main__':
    main()
