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
import threading

from common import *

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
    for i in range(5):
        start_one_kv_server(i)
    initconfig = ["0.0.0.0:12100", "0.0.0.0:12101", "0.0.0.0:12102"]

    # Wait for the system to (probably) be running ...
    time.sleep(1.0)
    # ... then tell the config server about the initial config.
    start_command(["go", "run", "./cmd/admin", "-conf", "0.0.0.0:12000",
                   "init"] + initconfig, cwd=simplepbdir)

def reconfig_system():
    time.sleep(10) # warmup
    time.sleep(30)
    newconfig = ["0.0.0.0:12102", "0.0.0.0:12103", "0.0.0.0:12104"]
    run_command(["go", "run", "./cmd/admin", "-conf", "0.0.0.0:12000",
                   "reconfig"] + newconfig, cwd=simplepbdir)

config = {}

def main():
    atexit.register(cleanup_procs)
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    global config

    config = {
        'read': 0,
        'write': 1.0,
        'keys': 100000,
        'clientcpus': '4-7',
        'configcpus': '0',
        'kv0cpus': '0',
        'kv1cpus': '1',
        'kv2cpus': '2',
        'kv3cpus': '0', # XXX: not used together with kv0
        'kv4cpus': '1', # XXX: not used together with kv1
    }

    start_multi_node_kv_system()
    threading.Thread(target=reconfig_system).start()
    a = goycsb_bench_inst('pbkv', 35, 70, 128, config['read'], config['write'], config['keys'], config['clientcpus'])
    # closed_lt('pbkv', 128, path.join(global_args.outdir, 'pb-kvs.jsons'), config['read'], config['write'], config['keys'], num_threads, config['clientcpus'])

    with open(path.join(global_args.outdir, 'pb_reconfig.dat'), 'a+') as outfile:
        ops_so_far = 0
        for e in a:
            outfile.write('{0},{1}\n'.format(e[0], 2*(e[1] - ops_so_far)))
            ops_so_far = e[1]

if __name__=='__main__':
    main()
