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
    "--outfile",
    help="output file for benchmark results",
    required=True,
    default=None,
)
parser.add_argument(
    "--reads",
    help="percentage of ops that are reads (between 0.0 and 1.0)",
    required=False,
    default=0.0,
)
parser.add_argument(
    "-e",
    "--errors",
    help="print stderr from commands being run",
    action="store_true",
)

global_args = parser.parse_args()
simplepbdir = ''
goycsbdir = ''

procs = []

# os.makedirs(global_args.outdir, exist_ok=True)
scriptdir = os.path.dirname(os.path.abspath(__file__))
simplepbdir = os.path.dirname(scriptdir)
goycsbdir = os.path.join(os.path.dirname(os.path.dirname(simplepbdir)), "go-ycsb")

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

def parse_ycsb_output(output):
    # look for 'Run finished, takes...', then parse the lines for each of the operations
    # output = output[re.search("Run finished, takes .*\n", output).end():] # strip off beginning of output

    # NOTE: sample output from go-ycsb:
    # UPDATE - Takes(s): 12.6, Count: 999999, OPS: 79654.6, Avg(us): 12434, Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000
    patrn = '(?P<opname>.*)\s+- Takes\(s\): (?P<time>.*), Count: (?P<count>.*), OPS: (?P<ops>.*), Avg\(us\): (?P<avg_latency>.*), Min\(us\):.*\n' # Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000'
    ms = re.finditer(patrn, output, flags=re.MULTILINE)
    a = dict()
    for m in ms:
        a[m.group('opname').strip()] = {'thruput': float(m.group('ops')), 'avg_latency': float(m.group('avg_latency')), 'raw': output[m.span()[0] : m.span()[1]]}
    return a

def goycsb_bench(kvname:str, threads:int, warmuptime:int, runtime:int, valuesize:int, readprop:float, updateprop:float, keys:int, extra_args=[], cooldown=0):
    """
    Returns a dictionary of the form
    { 'UPDATE': {'thruput': 1000, 'avg_latency': 12345', 'raw': 'blah'},...}
    """

    p = start_command([path.join(goycsbdir, './go-ycsb'),
                       'run', kvname,
                       '-P', path.join(simplepbdir, "bench", kvname + '_workload'),
                       '--threads', str(threads),
                       '--target', '-1',
                       '--interval', '100',
                       '-p', 'operationcount=' + str(2**32 - 1),
                       '-p', 'fieldlength=' + str(valuesize),
                       '-p', 'requestdistribution=uniform',
                       '-p', 'readproportion=' + str(readprop),
                       '-p', 'updateproportion=' + str(updateprop),
                       '-p', 'warmuptime=' + str(warmuptime), # TODO: increase warmup
                       '-p', 'recordcount=' + str(keys),
                       ] + extra_args, cwd=goycsbdir)

    if p is None:
        return ''

    to_parse = ""
    optypes_seen = 0
    num_optypes = 1 if readprop == 0.0 or updateprop == 0.0 else 2

    for stdout_line in iter(p.stdout.readline, ""):
        if stdout_line.find('Takes(s): {0}.'.format(runtime)) != -1:
          to_parse += stdout_line
          optypes_seen += 1
          if optypes_seen == num_optypes:
              break
    time.sleep(cooldown)
    p.stdout.close()
    p.terminate()
    return parse_ycsb_output(to_parse)

def parse_ycsb_output_totalops(output):
    # look for 'Run finished, takes...', then parse the lines for each of the operations
    # output = output[re.search("Run finished, takes .*\n", output).end():] # strip off beginning of output

    # NOTE: sample output from go-ycsb:
    # UPDATE - Takes(s): 12.6, Count: 999999, OPS: 79654.6, Avg(us): 12434, Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000
    patrn = '(?P<opname>.*) - Takes\(s\): (?P<time>.*), Count: (?P<count>.*), OPS: (?P<ops>.*), Avg\(us\): (?P<avg_latency>.*), Min\(us\):.*\n' # Min(us): 28, Max(us): 54145, 99th(us): 29000, 99.9th(us): 41000, 99.99th(us): 49000'
    ms = re.finditer(patrn, output, flags=re.MULTILINE)
    a = 0
    time = None
    for m in ms:
        a += int(m.group('count'))
        time = float(m.group('time'))
    return (time, a)

def goycsb_load(kvname:str, threads:int, valuesize:int, keys:int, extra_args=[]):
    run_command([path.join(goycsbdir, './go-ycsb'),
                 'load', kvname,
                 '-P', path.join(simplepbdir, "bench", kvname + '_workload'),
                 '--threads', str(threads),
                 '-p', 'fieldlength=' + str(valuesize),
                 '-p', 'recordcount=' + str(keys),
                 ] + extra_args, cwd=goycsbdir)
