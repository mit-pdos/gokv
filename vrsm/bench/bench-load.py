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
parser.add_argument('--recordcount', type=int, required=True)
args = parser.parse_args()

def run_command(args, cwd=None, shell=False):
    return subprocess.run(args, shell=shell, text=True, cwd=cwd)

def goycsb_load(kvname:str, threads:int, valuesize:int, keys:int, extra_args=[]):
    goycsbdir = os.path.expanduser("~/go-ycsb")
    simplepbdir = os.path.expanduser("~/gokv/simplepb/")

    run_command([path.join(goycsbdir, './go-ycsb'),
                 'load', kvname,
                 '-P', path.join(simplepbdir, "bench", kvname + '_workload'),
                 '--threads', str(threads),
                 '-p', 'fieldlength=' + str(valuesize),
                 '-p', 'recordcount=' + str(keys),
                 ] + extra_args, cwd=goycsbdir)

valuesize = 128

goycsb_load("pbkv", 400, valuesize, args.recordcount, ['-p', f"pbkv.configAddr=10.10.1.4:12000"])
