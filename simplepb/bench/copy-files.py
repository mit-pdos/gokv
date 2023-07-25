#!/usr/bin/env python3
from os import system as do

for i in list(range(8)):
    # copy stuff over
    do(f"cd ../../..; rsync -a gokv --exclude gokv/simplepb/durable node{i}:")
    do(f"cd ../../..; rsync -a go-ycsb node{i}:")
    print(f"finished copying to node{i}")
