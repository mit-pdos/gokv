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
import glob

from common import *

def num_threads(i):
    if i < 5:
        return i + 1
    elif i < 25:
        return (i - 4) * 5
    else:
        return 50 + (i - 24) * 50

def closed_lt(kvname, valuesize, outfilename, readprop, updateprop, recordcount, thread_fn, benchcpus):
    data = []
    i = 5
    last_good_index = 5
    peak_thruput = 0
    # last_thruput = 10000
    # last_threads = 10

    while True:
        if i > last_good_index + 5:
            break
        threads = thread_fn(i)

        a = goycsb_bench(kvname, threads, 10, valuesize, readprop, updateprop, recordcount, benchcpus)
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

def start_one_kv_server(i:int):
    start_command(many_cpus(["go", "run", "./cmd/kvsrv", "-filename",
                             "multi_kvserver" + str(i) + ".data", "-port",
                             str(12100 + i)],
                            config['kv' + str(i) + 'cpus']),
                  cwd=simplepbdir)

def cleanup_durable_dir():
    files = glob.glob(path.join(simplepbdir, "durable", "*.data"))
    r = run_command(["rm", "-f"] + files, cwd=simplepbdir)

def start_multi_node_kv_system():
    cleanup_durable_dir()
    start_config_server()
    for i in range(3):
        start_one_kv_server(i)
    initconfig = ["0.0.0.0:12100", "0.0.0.0:12101", "0.0.0.0:12102"]

    # Wait for the system to (probably) be running ...
    time.sleep(1.0)
    # ... then tell the config server about the initial config.
    start_command(["go", "run", "./cmd/admin", "-conf", "0.0.0.0:12000",
                   "init"] + initconfig, cwd=simplepbdir)

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
        'kv0cpus': '1',
        'kv1cpus': '2',
        'kv2cpus': '3',
    }

    start_multi_node_kv_system()
    # time.sleep(1000000)
    closed_lt('pbkv', 128, path.join(global_args.outdir, 'pb-kvs.jsons'), config['read'], config['write'], config['keys'], num_threads, config['clientcpus'])

if __name__=='__main__':
    main()
