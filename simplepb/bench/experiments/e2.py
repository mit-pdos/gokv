#!/usr/bin/env python3

# Recover+reconfiguration
#!/usr/bin/env python3

from os import system as do
import os
import time

print("""# Prints out the instantaneous latency and throughput of GroveKV with a
         # crash and then a reconfiguration in the middle""")

gobin='/usr/local/go/bin/go'
do(f"./start-pb.py --totalreplicas 4 --ncores 8 2 > /tmp/ephemeral.out 2>/tmp/ephemeral.err")

do("""~/go-ycsb/go-ycsb load pbkv -P /users/upamanyu/gokv/simplepb/bench/pbkv_workload --threads 200 --target -1 \
--interval 200 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion=1.0 \
-p updateproportion=1.0 -p warmuptime=5 -p recordcount=1000000 -p pbkv.configAddr=10.10.1.5:12000
""")

o = os.popen("""~/go-ycsb/go-ycsb run pbkv -P /users/upamanyu/gokv/simplepb/bench/pbkv_workload --threads 200 --target -1 \
--interval 200 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion=0.0 \
-p updateproportion=1.0 -p warmuptime=5 -p recordcount=1000000 -p pbkv.configAddr=10.10.1.5:12000 > /tmp/writes.txt \
""")

o = os.popen("""~/go-ycsb/go-ycsb run pbkv -P /users/upamanyu/gokv/simplepb/bench/pbkv_workload --threads 100 --target -1 \
--interval 200 -p operationcount=4294967295 -p fieldlength=128 -p requestdistribution=uniform -p readproportion=1.0 \
-p updateproportion=0.0 -p warmuptime=5 -p recordcount=1000000 -p pbkv.configAddr=10.10.1.5:12000 > /tmp/reads.txt \
""")

time.sleep(15)
print("Killing server")
do(f"""ssh upamanyu@node0 <<ENDSSH
cd /users/upamanyu/gokv/simplepb/;
killall kvsrv;
ENDSSH
""")

do(f"./reconfig.py 2 3") # servers are numbered starting at 0
time.sleep(15) # let it run for 10 seconds

do("killall go-ycsb")
do("./stop-pb.py")

# Analyze the file
# do("./get-inst-thruput.py")
