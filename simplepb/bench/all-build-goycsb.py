#!/usr/bin/env python3
from os import system as do

for i in [4, 5, 6, 7]:
    do(f"ssh node{str(i)} 'cd ~/go-ycsb; /usr/local/go/bin/go build ./cmd/go-ycsb'")
    print(f"done with node{str(i)}")
