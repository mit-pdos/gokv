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

parser = argparse.ArgumentParser(
description="Compute latency and throughput for kv service with varying client load"
)
parser.add_argument(
    "-n",
    "--dry-run",
    help="print commands without running them",
    action="store_true",
)
parser.add_argument(
    "-v",
    "--verbose",
    help="print commands in addition to running them",
    action="store_true",
)
parser.add_argument(
    "-e",
    "--errors",
    help="print stderr from commands being run",
    action="store_true",
)

procs = []
global_args = parser.parse_args()
scriptdir = os.path.dirname(os.path.abspath(__file__))
cmddir = scriptdir

resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))

def run_command(args, cwd=None, shell=False):
    if global_args.dry_run or global_args.verbose:
        print("[RUNNING] " + " ".join(args))
    if not global_args.dry_run:
        return subprocess.run(args, capture_output=True, shell=shell, text=True, cwd=cwd)

def start_command(args, cwd=None):
    if global_args.dry_run or global_args.verbose:
        print("[STARTING] " + " ".join(args))
    if not global_args.dry_run:
        e = subprocess.PIPE
        if global_args.errors:
            e = None
        p = subprocess.Popen(args, text=True, stdout=subprocess.PIPE, stderr=e, cwd=cwd, preexec_fn=os.setsid)
        global procs
        procs.append(p)
        return p

def cleanup_procs():
    global procs
    for p in procs:
        try:
            os.killpg(os.getpgid(p.pid), signal.SIGKILL)
        except Exception:
            continue
    procs = []

atexit.register(cleanup_procs)

def many_cores(args, c):
    return ["numactl", "-C", c] + args

def one_core(args, c):
    return ["numactl", "-C", str(c)] + args

def many_cpus(args, n):
    return ["numactl", "-N", n] + args


def main():
    # start n servers
    nservers = 5
    config = []
    for i in range(nservers):
        config.append('0.0.0.0:' + str(12000 + i))

    for i in range(nservers):
        start_command(['go', 'run', './srv', '-filename', 'nofile', '-port', str(12000 + i)] + config)

    time.sleep(100000)

if __name__=='__main__':
    main()
