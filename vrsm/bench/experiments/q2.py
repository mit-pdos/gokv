#!/usr/bin/env -S python3 -u
import os
from os import system as do
import argparse
from parseycsb import *

# Peak throughput with an increasing number of replica servers, each with 8 cores.

writeout = print

parser = argparse.ArgumentParser()
parser.add_argument('--threads', metavar='nthread', type=int,
                    help='number of client threads; default to 100 for "low load"',
                    default=100)
args = parser.parse_args()

os.chdir("/users/upamanyu/gokv/vrsm/bench")

do("./set-cores 8")
writeout(f"# peak throughput with increasing replicas")
writeout("nreplicas, throughput, threads, latency")
for nreplicas in [1, 2, 3, 4]:
    do(f"./start-pb.py --ncores 8 {nreplicas} > /tmp/ephemeral.out 2>/tmp/ephemeral.err")
    peakData = os.popen("./find-peak.py --onlypeak 2>/tmp/peak.err").read().rstrip()
    writeout(f"{nreplicas}, {peakData}")
