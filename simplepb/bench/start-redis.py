#!/usr/bin/env python3

from os import system as do
import time

ncores = 1

do(f"ssh upamanyu@node0 'killall go kvsrv config redis-server 2>/dev/null' ")
do(f"ssh upamanyu@node0 'rm -f /users/upamanyu/gokv/simplepb/durable/*'")
do(f"ssh upamanyu@node0 'cp /users/upamanyu/redis/redisraft/redisraft.so /users/upamanyu/gokv/simplepb/durable/'")

do(f""" ssh upamanyu@node0 <<ENDSSH
    cd /users/upamanyu/gokv/simplepb/;
    ./bench/set-cores.py {str(ncores)};
  nohup /users/upamanyu/redis/redis/src/redis-server \
        --port 5001 --dbfilename dbfilename, \
        --protected-mode no \
        --loadmodule ./redisraft.so \
        --raft.log-filename logfilename \
        --dir /users/upamanyu/gokv/simplepb/durable \
        --raft.log-fsync yes \
        --raft.addr 0.0.0.0:5001 1>/tmp/redis.out 2>/tmp/redis.err & ;
     sleep 2;
     /users/upamanyu/redis/redis/src/redis-cli -h 0.0.0.0 -p 5001 raft.cluster init
ENDSSH
""")

time.sleep(2)

