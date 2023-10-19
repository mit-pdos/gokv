#!/usr/bin/env python3
from os import system as do

for i in range(5):
    do(f"ssh node{str(i)} 'killall go kvsrv config redis-server 2>/dev/null' ")
    do(f"ssh node{str(i)} 'rm -f ~/gokv/vrsm/durable/kv.data ~/gokv/vrsm/durable/config.data'")
