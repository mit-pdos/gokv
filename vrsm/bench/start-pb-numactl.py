#!/usr/bin/env python3
from os import system as do
import time
import argparse
import sys

parser = argparse.ArgumentParser()
parser.add_argument('nreplicas', metavar='nreplicas', type=int,
                    help='number of replicas to set up and start')
parser.add_argument('--ncores', metavar='ncores', type=int,
                    default=8,
                    help='number of cores per replica server')
args = parser.parse_args()

do("./stop-pb.py")
print("Stopped pb (and deleted old files)")

gobin='/usr/local/go/bin/go'

nreplicas = args.nreplicas
ncores = args.ncores

if nreplicas > 3:
    print("too many replicas; can start at most 3")
    sys.exit(1)

# Start all replicas
for i in range(5):
    do(f"""ssh upamanyu@node{str(i)} <<ENDSSH
    cd /users/upamanyu/gokv/vrsm/;
    nohup {gobin} run ./cmd/kvsrv -filename kv.data -port 12100 1>/tmp/replica.out 2>/tmp/replica.err &
ENDSSH
    """)

# Start config server, on the next machine
do(f"""ssh upamanyu@node3 <<ENDSSH
    cd /users/upamanyu/gokv/vrsm/;
    nohup {gobin} run ./cmd/config -port 12000 1>/tmp/config.out 2>/tmp/config.err &
ENDSSH
    """)

time.sleep(2.0)
servers = ' '.join([f'10.10.1.{str(i + 1)}:12100' for i in range(nreplicas)])
do(f"go run ../cmd/admin -conf 10.10.1.4:12000 init {servers}")
