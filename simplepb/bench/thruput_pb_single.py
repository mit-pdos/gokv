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
    if i < 5:
        return 1 + i
    elif i < 15:
        return 10 + (i - 5) * 5
    else:
        return 100 + (i - 15) * 50

def closed_lt(kvname, warmuptime, runtime, valuesize, outfilename, readprop, updateprop, recordcount, thread_fn, cpuconfig):
    i = 0
    last_good_index = 0
    peak_thruput = 0
    while True:
        if i > last_good_index + 5:
            break
        threads = thread_fn(i)

        a = goycsb_bench(kvname, threads, warmuptime, runtime, valuesize, readprop, updateprop, recordcount, cpuconfig,
                         ['-p', f"pbkv.configAddr={config['serverhost']}:12000"])
        p = {'service': kvname, 'num_threads': threads, 'lts': a}

        with open(outfilename, 'a+') as outfile:
            outfile.write(json.dumps(p) + '\n')

        thput = sum([ a[op]['thruput'] for op in a ])

        if thput > peak_thruput:
            last_good_index = i
        if thput > peak_thruput:
            peak_thruput = thput

        i = i + 1

    return

gobin = "/usr/local/go/bin/go"
def start_config_server():
    # FIXME: core pinning
    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                             many_cpus([gobin, "run", "./cmd/config", "-port", "12000"], config['configcpus']), simplepbdir)))

def start_one_kv_server(kvcpuconfig):
    # delete kvserver.data file
    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                                           ["rm", "durable/single_kvserver.data"], simplepbdir))).wait()

    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                             many_cpus([gobin, "run", "./cmd/kvsrv", "-filename", "single_kvserver.data", "-port", "12100"], kvcpuconfig), simplepbdir)))

def start_single_node_kv_system(kvcpuconfig):
    start_config_server()
    start_one_kv_server(kvcpuconfig)
    time.sleep(2.0)
    # tell the config server about the initial config
    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                           [gobin, "run", "./cmd/admin", "-conf", "0.0.0.0:12000", "init", config['serverhost'] + ":12100"], simplepbdir))
                        ).wait()
    time.sleep(1.0)

config = {}

def main():
    atexit.register(cleanup_procs)
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    global config

    config = {
        'read': 0,
        'write': 1.0,
        'keys': 1000,
        'clientcpus': ['-C', '0-15'],
        'configcpus': ['-N', '0'],
        'kvcpuconfigs': [['-C', '0'],
                         ['-C', '0-1'],
                         ['-C', '0-2'],
                         ['-C', '0-3'],
                         ['-C', '0-4'],
                         ['-C', '0-5'],
                         ['-C', '0-6'],
                         ['-C', '0-7'],
                         ],
        'serverhost': '10.10.1.2',
        'warmuptime': 10,
        'runtime': 10,
    }

    for kvcpuconfig in config['kvcpuconfigs']:
        filename = datetime.now().strftime("%m-%d-%H-%M-%S") + "-pb-kvs.jsons"
        outfilepath = path.join(global_args.outdir, filename)

        os.system(f"ssh upamanyu@{config['serverhost']} 'killall go kvsrv config redis-server' ")
        start_single_node_kv_system(kvcpuconfig)
        with open(outfilepath, 'a+') as outfile:
            outfile.write(f"# Run with kvcpuconfig = {kvcpuconfig}\n")

        closed_lt('pbkv', config['warmuptime'], config['runtime'], 128, outfilepath, config['read'], config['write'], config['keys'], num_threads, config['clientcpus'])
        cleanup_procs()

if __name__=='__main__':
    main()
