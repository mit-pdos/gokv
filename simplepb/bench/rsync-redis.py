#!/usr/bin/env python3
from os import system as do

for i in [0, 4]:
    do(f"cd ../../..; rsync -a redis node{i}:")
