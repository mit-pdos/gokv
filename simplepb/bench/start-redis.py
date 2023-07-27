#!/usr/bin/env python3

from os import system as do
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('--ncores', metavar='ncores', type=int,
                    default=8,
                    help='number of cores per redis server')
args = parser.parse_args()

ncores = args.ncores

do(f"ssh node0 'killall go kvsrv config redis-server 2>/dev/null' ")
do(f"ssh node0 'rm -rf ~/gokv/simplepb/durable/*'")

do(f""" ssh node0 <<ENDSSH
    cd ~/gokv/simplepb/;
    ./bench/set-cores.py {str(ncores)};
    nohup ~/redis/redis/src/redis-server \
        --port 5001 --dbfilename dbfilename, \
        --protected-mode no \
        --dir ~/gokv/simplepb/durable \
        --appendonly yes \
        --appendfsync always \
        --save "" \
        1>/tmp/redis.out 2>/tmp/redis.err & ;
ENDSSH
""")
