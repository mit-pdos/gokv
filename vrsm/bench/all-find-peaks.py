#!/usr/bin/env python3

from os import system as do

for nreplicas in [1,2,3]:
    for ncores in range(1, 8):
        do(f"./start-pb.py {str(nreplicas)} --ncores {str(ncores)} > /dev/null")
        print(f"# running {str(nreplicas)} servers with {str(ncores)} cores each")
        do("./find-peak.py")
