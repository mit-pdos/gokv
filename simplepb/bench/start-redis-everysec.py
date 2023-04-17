#!/usr/bin/env python3

from os import system as do
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('--ncores', metavar='ncores', type=int,
                    default=8,
                    help='number of cores per redis server')
args = parser.parse_args()

ncores = args.ncores

do(f"ssh upamanyu@node0 'killall go kvsrv config redis-server 2>/dev/null' ")
do(f"ssh upamanyu@node0 'rm -f /users/upamanyu/gokv/simplepb/durable/*'")
do(f"ssh upamanyu@node0 'cp /users/upamanyu/gokv/simplepb/durable/'")

do(f""" ssh upamanyu@node0 <<ENDSSH
    cd /users/upamanyu/gokv/simplepb/;
    ./bench/set-cores.py {str(ncores)};
    nohup /users/upamanyu/redis/redis/src/redis-server \
        --port 5001  \
        --protected-mode no \
        --dir /users/upamanyu/gokv/simplepb/durable \
        --save "" \
        --appendonly yes\
        --appendfsync everysec 1>/tmp/redis.out 2>/tmp/redis.err & ;
ENDSSH
""")
