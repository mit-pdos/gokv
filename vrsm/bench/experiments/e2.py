#!/usr/bin/env python3

print("""# Runs GroveKV, and kills one of the servers and triggers reconfiguration in the
# middle. Outputs the instantaneous throughput of some read clients
# and some write clients during the reconfiguration.
# Data is put in `./data/reconfig`
""")

from os import system as do
import os
import time

os.chdir(os.path.expanduser('~/gokv/simplepb/bench'))
do(f"./start-pb.py --totalreplicas 4 --ncores 8 2 > /tmp/ephemeral.out 2>/tmp/ephemeral.err")

do("""~/go-ycsb/go-ycsb load pbkv -P ~/gokv/simplepb/bench/pbkv_workload --threads 200 --target -1 \
--interval 200 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion=1.0 \
-p updateproportion=1.0 -p warmuptime=5 -p recordcount=1000000 -p pbkv.configAddr=10.10.1.5:12000
""")

os.popen("""~/go-ycsb/go-ycsb run pbkv -P ~/gokv/simplepb/bench/pbkv_workload --threads 80 --target -1 \
--interval 200 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion=0.0 \
-p updateproportion=1.0 -p warmuptime=5 -p recordcount=1000000 -p pbkv.configAddr=10.10.1.5:12000 > /tmp/writes.txt \
""")

os.popen("""~/go-ycsb/go-ycsb run pbkv -P ~/gokv/simplepb/bench/pbkv_workload --threads 40 --target -1 \
--interval 200 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion=1.0 \
-p updateproportion=0.0 -p warmuptime=5 -p recordcount=1000000 -p pbkv.configAddr=10.10.1.5:12000 > /tmp/reads.txt \
""")

time.sleep(15)
print("Killing server")
do(f"""ssh node0 'killall kvsrv'""")

do(f"./reconfig.py 2 3") # servers are numbered starting at 0
time.sleep(15) # let it run for 10 seconds

do("killall go-ycsb")
do("./stop-pb.py")

# Analyze the file
do ("mkdir -p ./data/reconfig")
do("./get-inst-thruput.py /tmp/reads.txt > ./data/reconfig/reads.dat")
do("./get-inst-thruput.py /tmp/writes.txt > ./data/reconfig/writes.dat")
