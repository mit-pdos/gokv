#!/usr/bin/env python3
from os import system as do
import time
import argparse
import sys

parser = argparse.ArgumentParser()
parser.add_argument('nreplicas', metavar='nreplicas', type=int,
                    help='number of replicas to set up and start')
parser.add_argument('--totalreplicas', type=int, default=3,
                    help='max number of replica servers (config service is started on next server after this max)')
parser.add_argument('--ncores', metavar='ncores', type=int,
                    default=8,
                    help='number of cores per replica server')
args = parser.parse_args()

do("./stop-pb.py")
print("Stopped pb (and deleted old files)")

gobin='/usr/local/go/bin/go'

totalreplicas = args.totalreplicas
nreplicas = args.nreplicas
ncores = args.ncores

if nreplicas > totalreplicas:
    print(f"too many replicas; can start at most {totalreplicas}")
    sys.exit(1)

servers = ' '.join([f'10.10.1.{str(i + 1)}:12100' for i in range(nreplicas)])
# Start config server, on the last machine that isn't the client, with all 8 cores
do(f"""ssh node{totalreplicas} <<ENDSSH
    cd ~/gokv/vrsm/;
    ./bench/set-cores.py 8;
    nohup {gobin} run ./cmd/config -port 12000 {servers} 1>/tmp/config.out 2>/tmp/config.err &
ENDSSH
    """)

time.sleep(1) # make sure config service is up
conf_addr = f"10.10.1.{totalreplicas + 1}:12000"
# Start all replicas
for i in range(totalreplicas):
    do(f"""ssh node{str(i)} <<ENDSSH
    cd ~/gokv/vrsm/;
    ./bench/set-cores.py {ncores};
    nohup {gobin} run ./cmd/kvsrv -conf {conf_addr} -filename kv.data -port 12100 1>/tmp/replica.out 2>/tmp/replica.err &
ENDSSH
    """)

time.sleep(2.0)
do(f"go run ../cmd/mkleader -host 10.10.1.{totalreplicas + 1}:12001")
do(f"go run ../cmd/admin -conf 10.10.1.{totalreplicas + 1}:12000 init {servers}")
