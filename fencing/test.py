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

parser = argparse.ArgumentParser()
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
global_args = parser.parse_args()

procs = []
def run(cmd):
    args = args.split()
    if global_args.dry_run or global_args.verbose:
        print("[RUNNING] " + " ".join(args))
    if not global_args.dry_run:
        return subprocess.run(args, capture_output=True, text=True)

def start(cmd, cwd=None):
    args = args.split()
    if global_args.dry_run or global_args.verbose:
        print("[STARTING] " + " ".join(args))
    if not global_args.dry_run:
        p = subprocess.Popen(args, text=True, stdout=subprocess.PIPE, cwd=cwd)
        global procs
        procs.append(p)
        return p

def cleanup_procs():
    for p in procs:
        p.kill()

def start_config(port=12000):
    return start("go run ./cmd/config -port 12000")

def start_client(config_ip="127.0.0.1", config_port=12000):
    return start(f"go run ./cmd/loopclient -config {config_ip}:{str(config_port)}")

def start_frontend(my_port, config_ip="127.0.0.1", config_port=12000):
    return start(f"go run ./cmd/loopclient -config {config_ip}:{str(config_port)} -port {my_port}")

# Start one config server,
# 2 front-end servers,
# 1 client.
#
# Kill the first front-end in the middle of the test.

def test_simple_failover():
    start_config()
    frontend0 = frontends.append(start_frontend(12100))
    cl = start_client()
    time.sleep(0.5)

    frontend1 = frontends.append(start_frontend(12101))
    frontend0.kill()

    cleanup_procs()
