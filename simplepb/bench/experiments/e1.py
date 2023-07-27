#!/usr/bin/env python3

# Generates raw data for latency/throughput curves for Redis and GroveKV for
# various read ratios.
# Puts final data in `gokv/simplepb/bench/data/redis_vs_grove/`.
# This is intended to help find the peak throughput of the two systems and the
# latency under low load.

from os import system as do
import os
import json
import sys
from bench import lt_pb_single, lt_redis_single
import signal

from statistics import stdev, mean
import numpy as np
import numpy as np, scipy.stats as st

def sigint_handler(sig, frame):
    sys.exit(0)
signal.signal(signal.SIGINT, sigint_handler)

os.chdir(os.path.expanduser('~/gokv/simplepb/bench'))

def get_peak(data):
    maxthruput = 0
    maxthreads = 0
    for d in data:
        thru = float(d['lts']['TOTAL']['thruput'])
        if thru > maxthruput:
            maxthruput = thru
            maxthreads = int(d['num_threads'])
    return maxthruput, maxthreads

usecache = True
def load_cache(fname):
    if not usecache:
        return None
    try:
        with open(fname, 'r') as f:
            return json.loads(f.read())
    except FileNotFoundError as e:
        pass
    except Exception as e:
        print("load_cache: ", e)
        pass
    return None

def save_cache(o, fname):
    with open(fname, 'w') as f:
        f.write(json.dumps(o))

def with_cache(name, f):
    o = load_cache(name)
    if o is None:
        o = f()
        save_cache(o, name)
    return o

def num_threads(i):
    if i < 1:
        return 1
    i = i - 1
    if i < 10:
        return 5 + i * 5
    i = i - 10
    return 100 * (i + 1)

def num_threads_lite(i):
    return 400 + 200 * i


outfilename = "./data/redis_vs_grove/peak-table-lite.tex"
with open(outfilename, 'a+') as f:
    pass

peak_threads = []
for readratio in [0.0, 0.5, 0.95]:
    do('mkdir /tmp/gokv')
    do('mv /tmp/gokv/grovekv-lts.txt /tmp/grovekv-lts.old')
    do('mv /tmp/gokv/redis-lts.txt /tmp/redis-lts.old')

    id = str(int(100 * readratio))
    # do(f'./lt_pb_single.py -v -e --reads {str(readratio)} --outfile /tmp/gokv/grovekv-lts.txt 1>/tmp/pb.out 2>/tmp/pb.err')
    # do(f'./lt_redis_single.py -v -e --reads {str(readratio)} --outfile /tmp/gokv/redis-lts.txt 1>/tmp/redis.out 2>/tmp/redis.err')
    grove_data = with_cache(f"./data/redis_vs_grove/grovekv-lts-{id}.txt",
                            lambda: lt_pb_single.run("/tmp/gokv/grovekv-lts.txt",
                                                        readratio, num_threads_lite,
                                                        warmuptime=10, runtime=20))

    redis_data = with_cache(f"./data/redis_vs_grove/redis-lts-{id}.txt",
                            lambda: lt_redis_single.run("/tmp/gokv/redis-lts.txt",
                                                        readratio, num_threads_lite,
                                                        warmuptime=10, runtime=20))

    peak_grove, threads_grove = get_peak(grove_data)
    peak_redis, threads_redis = get_peak(redis_data)

    peak_threads.append((readratio, threads_grove, threads_redis))

    with open(outfilename, 'a+') as f:
        print(f"Throughput for YCSB {str(100 - int(100*readratio))}\\% writes & {peak_redis}~req/s & {peak_grove}~req/s \\\\",
            file=f)


# Get latency at low load for both reads and writes for both systems
def single_thread_fn(n):
    def f(i):
        if i > 0:
            return None
        return n
    return f

def get_latencies():
    grove_read = with_cache("./data/redis_vs_grove/grove_read.txt",
                            lambda:
                            lt_pb_single.run("/tmp/gokv/grovekv-lts.txt",
                                             1.0, single_thread_fn(1),
                                             warmuptime=30,
                                             runtime=120)[0]['lts']['TOTAL']['avg_latency'])

    grove_write = with_cache("./data/redis_vs_grove/grove_write.txt",
                             lambda:
                             lt_pb_single.run("/tmp/gokv/grovekv-lts.txt",
                                              0.0, single_thread_fn(1),
                                              warmuptime=30,
                                              runtime=120)[0]['lts']['TOTAL']['avg_latency'])

    redis_read = with_cache("./data/redis_vs_grove/redis_read.txt",
                            lambda:
                            lt_redis_single.run("/tmp/gokv/redis-lts.txt",
                                                1.0, single_thread_fn(1),
                                                warmuptime=30,
                                                runtime=120)[0]['lts']['TOTAL']['avg_latency'])

    redis_write = with_cache("./data/redis_vs_grove/redis_write.txt",
                             lambda:
                             lt_redis_single.run("/tmp/gokv/redis-lts.txt",
                                                 0.0, single_thread_fn(1),
                                                 warmuptime=30,
                                                 runtime=120)[0]['lts']['TOTAL']['avg_latency'])

    with open("./data/redis_vs_grove/latency.tex", "a+") as f:
        print(f"Read latency under low load & {redis_read}~us & {grove_read}~us \\\\",
              file=f)

        print(f"Write latency under low load & {redis_write}~us & {grove_write}~us \\\\",
              file=f)

get_latencies()

outfilename = "./data/redis_vs_grove/peak-table.tex"
with open(outfilename, 'a+') as f:
    pass
num_samples = 10

def get_stats(samples):
    thrus = []
    for d in samples:
        thrus.append(float(d['lts']['TOTAL']['thruput']))

    i = st.t.interval(confidence=0.95, df=len(thrus)-1,
                      loc=np.mean(thrus),
                      scale=st.sem(thrus))
    return np.mean(thrus), i

# Next, run the thread count that resulted in peak throughput multiple times and
# for longer to get some statistics on peak throughput.
for (readratio, threads_grove, threads_redis) in peak_threads:
    id = str(int(100 * readratio))

    def get_grove_samples():
        samples = []
        for i in range(num_samples):
            samples += lt_pb_single.run("/tmp/gokv/grovekv-lts.txt",
                                        readratio, single_thread_fn(threads_grove),
                                        warmuptime=30, runtime=120)
        return samples

    grove_samples = with_cache(f"./data/redis_vs_grove/grovekv-{id}-samples.txt", get_grove_samples)
    grove_avg, grove_ci = get_stats(grove_samples)

    def get_redis_samples():
        samples = []
        for i in range(num_samples):
            samples += lt_redis_single.run("/tmp/gokv/redis-lts.txt",
                                           readratio, single_thread_fn(threads_redis),
                                           warmuptime=30, runtime=120)
        return samples

    redis_samples = with_cache(f"./data/redis_vs_grove/redis-{id}-samples.txt", get_redis_samples)
    redis_avg, redis_ci = get_stats(redis_samples)

    with open(outfilename, 'a+') as f:
        print(f"% YCSB {str(100 - int(100*readratio))}\\% writes stats: {redis_avg,redis_ci} & {grove_avg,grove_ci} \\\\",
            file=f)
        print(f"Throughput for YCSB {str(100 - int(100*readratio))}\\% writes & {redis_avg}~req/s & {grove_avg}~req/s \\\\",
            file=f)


# Example output:
# Throughput for YCSB 100\% writes & 68,594~req/s & 72,005~req/s \\
# Throughput for YCSB 50\% writes & 81,821~req/s & 83,965~req/s \\
# Throughput for YCSB 5\% writes & 100,732~req/s & 91,265~req/s \\
#
# Read latency under low load & 178~us & 128~us \\
# Write latency under low load & 570~us & 571~us \\
