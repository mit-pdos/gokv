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
import threading

from config import *

parser = argparse.ArgumentParser(
description="Find peak throughput of KV service for a varying number of shard servers"
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
global_args = parser.parse_args()
gokvdir = ''

procs = []

def run_command(args, cwd=None):
    if global_args.dry_run or global_args.verbose:
        print("[RUNNING] " + " ".join(args))
    if not global_args.dry_run:
        return subprocess.run(args, capture_output=True, text=True, cwd=cwd)

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

def many_cores(args, c):
    return ["numactl", "-C", c] + args

def one_core(args, c):
    return ["numactl", "-C", str(c)] + args

def start_system():
    confhost = "127.0.0.1:" + config["confport"]
    conf = []
    for p in config["replicaports"]:
        conf = conf + ["127.0.0.1:" + p]

    # FIXME: start this
    # start_command(["go", "run", "./cmd/confsrv", "-port", config["confport"]])

    for p in config["replicaports"]:
        start_command(["go", "run", "./cmd/replicasrv", "-port", p])

    time.sleep(2)
    # wait for servers to be up so that we can start the controller and have the
    # controller inform the primary about the current config
    start_command(["go", "run", "./cmd/ctrlsrv", # "-conf", confhost,
                   "-port", config["ctrlport"]] + conf)
    return

def main():
    atexit.register(cleanup_procs)
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))

    start_system()
    time.sleep(9999999)

if __name__=='__main__':
    main()
