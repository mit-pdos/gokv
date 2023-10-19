#!/usr/bin/env python3
from os import system as do

for i in range(5):
    # copy stuff over
    do(f"cd ../../..; rsync -a gokv --exclude gokv/vrsm/durable node{i}:")
    do(f"cd ../../..; rsync -a go-ycsb node{i}:")
    # do(f"ssh node{i}: 'cd gokv/vrsm/bench; ./lt_pb_single.py'")
