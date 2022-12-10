#!/usr/bin/env python3

# Run one primary and one backup.
# Backup has 8 cores, primary has a varying number (1,2,...,8).
# Measures the peak throughput

import os
from os import system as do
import time
import sys

gobin='/usr/local/go/bin/go'
os.chdir('/users/upamanyu/gokv/simplepb/bench')

def start_primary(ncores):
    do(f"""ssh upamanyu@node0 <<ENDSSH
        cd /users/upamanyu/gokv/simplepb/;
        ./bench/set-cores.py {ncores};
        nohup {gobin} run ./cmd/kvsrv -filename kv.data -port 12100 1>/tmp/replica.out 2>/tmp/replica.err &
ENDSSH
""")

def start_backup():
    do(f"""ssh upamanyu@node1 <<ENDSSH
        cd /users/upamanyu/gokv/simplepb/;
        ./bench/set-cores.py 8;
        nohup {gobin} run ./cmd/kvsrv -filename kv.data -port 12100 1>/tmp/replica.out 2>/tmp/replica.err &
ENDSSH
        """)

def start_config():
    do(f"""ssh upamanyu@node3 <<ENDSSH
        cd /users/upamanyu/gokv/simplepb/;
        ./bench/set-cores.py 8;
        nohup {gobin} run ./cmd/config -port 12000 1>/tmp/config.out 2>/tmp/config.err &
ENDSSH
        """)

def restart_system(primarycores):
    do("./stop-pb.py")

    start_primary(primarycores)
    start_backup()
    start_config()

    servers = ' '.join(['10.10.1.1:12100', '10.10.1.2:12100'])
    time.sleep(2.0)
    do(f"go run ../cmd/admin -conf 10.10.1.4:12000 init {servers}")
    return

for nprimarycores in range(1,9):
    restart_system(nprimarycores)
    # find peak throughput at this configuration
    with open("./data/peak.txt", "a+") as f:
        f.write(f"# Running with {nprimarycores} cores on primary, 8 on backup\n")

    do("./find-peak.py 2>/tmp/peak.err >> ./data/peak.txt")
