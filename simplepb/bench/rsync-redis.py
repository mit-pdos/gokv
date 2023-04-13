#!/usr/bin/env python3
from os import system as do

for i in range(1):
    do(f"cd ../../..; rsync -a redis node{i}:")
