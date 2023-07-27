#!/usr/bin/env -S python3 -u

import os
from os import system as do
from .parseycsb import *
import sys
import argparse
import json
from .bench_ops_multiclient import *

def restart_system(ncores, nreplicas):
    do_quiet("./stop-pb.py")
    do_quiet(f"./start-pb.py --ncores {ncores} {nreplicas}")
    do_quiet(f"./bench-load.py --recordcount {recordcount}")
    return

client_machines = [5, 6, 7]
recordcount = 1000

def find_peak(ncores, nreplicas, reads, threadcounts, runtime=40, cooldown=10, warmuptime=20, outfilename=None):
    highestThreads = 0
    highestThruput = 0
    for threads in threadcounts:
        restart_system(ncores, nreplicas)
        ms = get_thruput(client_machines, threads, reads, recordcount, runtime, cooldown, warmuptime)

        wthruput = 0.0
        wlatency = 0.0
        rthruput = 0.0
        rlatency = 0.0
        for m in ms:
            if 'UPDATE' in m["lts"]:
                wthruput += m["lts"]['UPDATE']["thruput"]
                wlatency += m["lts"]['UPDATE']["avg_latency"]
            if 'READ' in m["lts"]:
                rthruput += m["lts"]['READ']["thruput"]
                rlatency += m["lts"]['READ']["avg_latency"]

        rlatency = rlatency/len(ms)
        wlatency = wlatency/len(ms)

        if outfilename:
            with open(outfilename, 'a+') as f:
                print(f"{threads}, {wthruput}, {wlatency}, {rthruput}, {rlatency}", file=f)

        thruput = rthruput + wthruput

        if thruput > highestThruput:
            highestThruput = thruput
            highestThreads = threads
        if 1.50 * thruput < highestThruput:
            break # don't bother trying them all

    return highestThruput, highestThreads

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--onlypeak',
                        help='if this flag is set, then this only prints the peak throughput, not all the intermediate throughputs',
                        action='store_true')
    parser.add_argument("--reads",
                        help="percentage of ops that are reads (between 0.0 and 1.0)",
                        required=False,
                        default=0.0)
    parser.add_argument("--ncores", help="number of cores per server", required=True)
    parser.add_argument("--nreplicas", help="number of servers", required=True)
    parser.add_argument('--benchcmd',
                        help='The command to run to benchmark a single server (should be either the one for GroveKV or for redis)',
                        default='./bench-ops-multiclient.py')
    args = parser.parse_args()

    threadcounts = [50, 100, 150, 200, 250, 400, 600] # + [200 * (i + 2) for i in range(3)] + [500 * (i + 3) for i in range (10)]
    runtime = 40
    warmuptime = 20
    cooldown = 10
    fullwarmuptime = 20
    fullruntime = 120
