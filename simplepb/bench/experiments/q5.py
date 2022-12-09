#!/usr/bin/env python3

from os import system as do
import os
import time

print("""# Prints out the instantaneous latency and throughput of GroveKV with a
         # crash and then a reconfiguration in the middle""")

gobin='/usr/local/go/bin/go'
for ncores in range(1,9):
    do(f"./start-pb.py 2 --ncores 8 > /tmp/ephemeral.out 2>/tmp/ephemeral.err")

    o = os.popen("./bench-put.py --interval 1000 --warmup 0 100 1>/tmp/reconfig_raw.txt")
    time.sleep(10) # let it run for 10 seconds

    print("Killing server")
    # kill then restart the server.
    do(f"""ssh upamanyu@node1 <<ENDSSH
    cd /users/upamanyu/gokv/simplepb/;
    killall kvsrv;
    nohup {gobin} run ./cmd/kvsrv -filename kv.data -port 12100 1>/tmp/replica.out 2>/tmp/replica.err &
ENDSSH
    """)

    time.sleep(10) # let it run for 10 seconds
    do(f"./reconfig.py 1 2") # servers are numbered starting at 0, so this is turning the old backup into a primary and adding a new server
    time.sleep(20) # let it run for another 10 seconds
    o.close() # exit benchmark
    do("./stop-pb.sh")
    sys.exit(0)
