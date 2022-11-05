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
        return (i - 4) * 5
    else:
        return 50 + (i - 24) * 50

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

        # FIXME: clear out redis data
        cleanup_procs()
        start_fresh_single_node_redisraft()
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

config = {}

def start_fresh_single_node_redisraft():
    durable_dir = os.path.join(simplepbdir, 'durable')
    dbfilename = 'raft1.rdb'
    logfilename = 'raftlog1.db'

    # clean up old files
    run_command(["rm", dbfilename, logfilename, logfilename + ".meta", logfilename + ".idx"], cwd=durable_dir)

    run_command(["cp", os.path.join(redisdir, "redisraft", "redisraft.so"), durable_dir])
    start_command(many_cpus(["./redis/src/redis-server",
                             "--port", "5001", "--dbfilename", dbfilename,
                             "--loadmodule", "./redisraft.so",
                             "--raft.log-filename", logfilename,
                             "--dir", durable_dir,
                             "--raft.log-fsync", "yes",
                             "--raft.addr", "localhost:5001",], config['rediscpus']), cwd=redisdir)

    time.sleep(1)
    run_command(["./redis/src/redis-cli", "-p", "5001", "raft.cluster", "init"], cwd=redisdir)
    time.sleep(2)

redisdir = ''

def main():
    atexit.register(cleanup_procs)
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    global config
    global redisdir
    redisdir = os.path.join(os.path.dirname(goycsbdir), 'redis')

    config = {
        'read': 0,
        'write': 1.0,
        'keys': 1000,
        'clientcpus': '4-7',
        # 'clientcpus': '0',
        'rediscpus': '0',
    }

    # start_fresh_single_node_redisraft()
    closed_lt('rediskv', 128, path.join(global_args.outdir, 'redis-kvs.jsons'), config['read'], config['write'], config['keys'], num_threads, config['clientcpus'])

if __name__=='__main__':
    main()
