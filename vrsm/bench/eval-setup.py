#!/usr/bin/env python3
from os import system as do
import os

# Set up SSH key between the cloudlab machines and downloads some packages
# (e.g. Go) on them. Also turn off hyperthreading.
do("./setup-machines.py")

# Copy the GroveKV and go-ycsb to all the cloudlab machines.
do("./copy-files.py")

# Copy redis to node0, which is the only machine that runs redis.
do("./rsync-redis.py")

# Build go-ycsb so it's ready for later for the client machines.
do("./all-build-goycsb.py")

# The cloudlab d430 machines have 2 separate CPUs, and to avoid strange NUMA
# effects when benchmarking, this turns off the second CPUs.
do("./all-only-one-cpu.py")

# Build redis now so it's ready to run later.
do(f"ssh node0 'cd gokv/simplepb/bench; ./build-redis.py'")
