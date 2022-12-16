#!/usr/bin/env python3
from os import system as do

for i in reversed(range(3)):
    # copy stuff over
    do(f"cd ../../..; rsync -a gokv --exclude gokv/simplepb/durable node{i}:")
    do(f"cd ../../..; rsync -a go-ycsb node{i}:")
    print(f"finished copying to node{i}")
    # do(f"ssh node{i}: 'cd gokv/simplepb/bench; ./lt_pb_single.py'")
