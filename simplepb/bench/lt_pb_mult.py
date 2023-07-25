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

def num_threads(i):
    if i < 10:
        return 5 + i * 5
    i = i - 10
    return 100 * (i + 1)

def reset_kv_system():
    os.system("./start-pb.py --ncores 8 3")

def run(outfilepath, readratio, warmuptime=30, runtime=120):
    resource.setrlimit(resource.RLIMIT_NOFILE, (100000, 100000))
    serverhost = '10.10.1.4'
    config = {
        'outfilename': outfilepath,
        'reads': readratio,
        'writes': 1 - readratio,
        'recordcount': 1000,
        'warmuptime': warmuptime,
        'runtime': runtime,
        'valuesize': 128,
    }

    closed_lt('pbkv', config, reset_kv_system, num_threads,
              ['-p', f"pbkv.configAddr={serverhost}:12000"])
    cleanup_procs()
