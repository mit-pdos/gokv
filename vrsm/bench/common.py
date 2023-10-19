#!/usr/bin/env python3
import argparse
import subprocess
import os
import resource
import atexit
import signal

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

def start_shell_command(cmd, cwd=None):
    if global_args.dry_run or global_args.verbose:
        print("[STARTING] " + cmd)
    if not global_args.dry_run:
        e = subprocess.PIPE
        if global_args.errors:
            e = None
        p = subprocess.Popen(cmd, text=True, stdout=subprocess.PIPE, shell=True, stderr=e, cwd=cwd, preexec_fn=os.setsid)
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

def many_cpus(args, c):
    return ["numactl"] + c + args


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

global_args = parser.parse_args()
vrsmdir = ''
goycsbdir = ''

procs = []
atexit.register(cleanup_procs)

# os.makedirs(global_args.outdir, exist_ok=True)
scriptdir = os.path.dirname(os.path.abspath(__file__))
vrsmdir = os.path.dirname(scriptdir)
goycsbdir = os.path.join(os.path.dirname(os.path.dirname(vrsmdir)), "go-ycsb")

resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
