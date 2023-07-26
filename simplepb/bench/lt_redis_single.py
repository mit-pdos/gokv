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
from datetime import datetime
from .goycsb import *

def reset_redis():
    os.system("./start-redis.py --ncores 1 1>/tmp/redis.out 2>/tmp/redis.err")

def run(outfilepath, readratio, threads_fn, warmuptime=30, runtime=120):
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    serverhost = '10.10.1.1'
    config = {
        'outfilename': outfilepath,
        'reads': readratio,
        'writes': 1 - readratio,
        'recordcount': 1000,
        'warmuptime': warmuptime,
        'runtime': runtime,
        'valuesize': 128,
    }

    data = closed_lt('rediskv', config, reset_redis, threads_fn,
              ['-p', f"redis.addr={serverhost}:5001"])
    cleanup_procs()
    return data
