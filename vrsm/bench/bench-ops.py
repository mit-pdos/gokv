#!/usr/bin/env python3
import os
from os import system as do
import argparse
import re
import time
import subprocess
import json
import signal
from os import path
import atexit

parser = argparse.ArgumentParser()
parser.add_argument("-n", "--dry-run", help="print commands without running them",
                    action="store_true",)
parser.add_argument("-v", "--verbose", help="print commands in addition to running them",
                    action="store_true",)
parser.add_argument('threads', metavar='nthread', type=int,
                    help='number of threads')
parser.add_argument('--warmup', type=int, default=10)
parser.add_argument('--cooldown', type=int, default=10)
parser.add_argument('--runtime', type=int, default=10)
parser.add_argument("--reads", help="percentage of ops that are reads (between 0.0 and 1.0)",
                    required=False, default=0.0)
parser.add_argument('--recordcount', type=int, default=1000)
parser.add_argument("--outfilename", help="the file where the output is placed", required=True)
args = parser.parse_args()

procs = []
def start_command(cmd, cwd=None):
    if args.dry_run or args.verbose:
        print("[STARTING] " + " ".join(cmd))
    if not args.dry_run:
        e = subprocess.PIPE
        if False: # args.errors:
            e = None
        p = subprocess.Popen(cmd, text=True, stdout=subprocess.PIPE, stderr=e, cwd=cwd, preexec_fn=os.setsid)
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

    goycsbdir = os.path.expanduser("~/go-ycsb/")
    simplepbdir = os.path.expanduser("~/gokv/simplepb/")

    # gobin = '/usr/local/go/bin/go'
    # run_command([gobin, 'build', './cmd/go-ycsb'], cwd=goyscbdir)
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
    num_optypes = 2 if readprop == 0.0 or updateprop == 0.0 else 3

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

valuesize = 128

a = goycsb_bench("pbkv", args.threads, args.warmup, args.runtime, valuesize,
                 float(args.reads), 1-float(args.reads), args.recordcount,
                 ['-p', f"pbkv.configAddr=10.10.1.4:12000"], cooldown=args.cooldown)
p = {'service': "pbkv", 'num_threads': args.threads, 'lts': a}
with open(args.outfilename, 'w') as f:
    print(json.dumps(p) + '\n', file=f)
