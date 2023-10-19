#!/usr/bin/env -S python3 -u
import os
from os import system as do
import argparse
from parseycsb import *

# Latency at medium load with increasing number of replica servers; up to 4
# replicas.

writeout = print

parser = argparse.ArgumentParser()
parser.add_argument('--threads', metavar='nthread', type=int,
                    help='number of client threads; default to 100 for "low load"',
                    default=100)
args = parser.parse_args()

warmup = 10
runtime = 10

os.chdir("/users/upamanyu/gokv/vrsm/bench")
do("./set-cores 8")
writeout(f"# average latency with {args.threads} client threads and increasing replicas ")
writeout("nreplicas, avglatency")
for nreplicas in [1, 2, 3, 4]:
    do(f"./start-pb.py --ncores 8 {nreplicas} > /tmp/ephemeral.out 2>/tmp/ephemeral.err")
    benchoutput = os.popen(f"./bench-put.py {args.threads} --warmup {warmup}", 'r', 100)
    ret = ''
    for line in benchoutput:
        if line.find('Takes(s): {0}.'.format(runtime)) != -1:
            ret = line
            benchoutput.close()
            break

    do("killall go-ycsb > /dev/null")
    time, ops, latency = (parse_ycsb_output_totalops(ret))
    writeout(f"{nreplicas}, {latency}")
