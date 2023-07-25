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
do(f"ssh node0 'rm -f ~/gokv/simplepb/durable/*'")

# do(f"ssh upamanyu@node0 'cp /users/upamanyu/redis/redisraft/redisraft.so /users/upamanyu/gokv/simplepb/durable/'")
# do(f""" ssh upamanyu@node0 <<ENDSSH
    # cd /users/upamanyu/gokv/simplepb/;
    # ./bench/set-cores.py {str(ncores)};
    # nohup /users/upamanyu/redis/redis/src/redis-server \
        # --port 5001 --dbfilename dbfilename, \
        # --protected-mode no \
        # --loadmodule ./redisraft.so \
        # --raft.log-filename logfilename \
        # --dir /users/upamanyu/gokv/simplepb/durable \
        # --raft.log-fsync yes \
        # --raft.addr 0.0.0.0:5001 1>/tmp/redis.out 2>/tmp/redis.err & ;
     # sleep 2;
     # /users/upamanyu/redis/redis/src/redis-cli -h 0.0.0.0 -p 5001 raft.cluster init;
     # sleep 1;
# ENDSSH
# """)

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
