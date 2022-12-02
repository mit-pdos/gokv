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
        return i + 1
    elif i < 25:
        return 5 + (i - 5) * 5
    else:
        return 500 + (i - 25) * 500

def closed_lt(kvname, warmuptime, runtime, valuesize, outfilename, readprop, updateprop, recordcount, thread_fn, benchcpus):
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
        a = goycsb_bench(kvname, threads, warmuptime, runtime, valuesize, readprop, updateprop, recordcount, benchcpus,
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
    durable_dir = os.path.join(simplepbdir, 'durable')
    dbfilename = 'raft1.rdb'
    logfilename = 'raftlog1.db'

    # clean up old processes
    os.system(f"ssh upamanyu@{config['serverhost']} 'killall go kvsrv config redis-server'")

    # clean up old files
    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                                            ["rm", "-f", dbfilename,
                                             logfilename, logfilename + ".meta",
                                             logfilename + ".idx"], durable_dir)
                                 )).wait()


    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                                            ["cp", os.path.join(redisdir, "redisraft", "redisraft.so"), durable_dir], cwd=redisdir))).wait()
    time.sleep(4)

    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                                           many_cpus(["./redis/src/redis-server",
                                                      "--port", "5001", "--dbfilename", dbfilename,
                                                      "--protected-mode", "no",
                                                      "--loadmodule", "./redisraft.so",
                                                      "--raft.log-filename", logfilename,
                                                      "--dir", durable_dir,
                                                      "--raft.log-fsync", "yes",
                                                      "--raft.addr", "0.0.0.0:5001",], config['rediscpus']),
                                           redisdir))
                        )

    time.sleep(2)
    start_shell_command(' '.join(remote_cmd(config['serverhost'],
                                            ["./redis/src/redis-cli",
                                             "-h", config['serverhost'],
                                             "-p",
                                             "5001", "raft.cluster", "init"],
                                            cwd=redisdir))).wait()
    time.sleep(1)

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
        'clientcpus': ['-C', '0-7'],
        'rediscpus': ['-C', '0'],
        'serverhost': '10.10.1.2',
        'warmuptime': 10,
        'runtime': 10,
    }

    filename = datetime.now().strftime("%m-%d-%H-%M-%S") + "-redis-kvs.jsons"
    outfilepath = path.join(global_args.outdir, filename)

    closed_lt('rediskv', config['warmuptime'], config['runtime'], 128, outfilepath, config['read'], config['write'], config['keys'], num_threads, config['clientcpus'])

if __name__=='__main__':
    main()
